package pgxpoolgo

import (
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"reflect"
	"strings"
)

/*
The codes below is based on `pgxmock` from `github.com/pashagolub/pgxmock`
*/

type mockRowSet struct {
	set *MockRow
	pos int
}

func (rs *mockRowSet) Scan(dest ...interface{}) error {
	r := rs.set
	if len(dest) != len(r.defs) {
		return fmt.Errorf("incorrect argument number %d for columns %d", len(dest), len(r.defs))
	}
	for i, col := range r.row {
		if dest[i] == nil {
			continue
		}
		destVal := reflect.ValueOf(dest[i])
		if destVal.Kind() != reflect.Ptr {
			return fmt.Errorf("destination argument must be a pointer for column %s", r.defs[i].Name)
		}
		if col == nil {
			dest[i] = nil
			continue
		}
		val := reflect.ValueOf(col)
		if _, ok := dest[i].(*interface{}); ok || destVal.Elem().Kind() == val.Kind() {
			if destElem := destVal.Elem(); destElem.CanSet() {
				destElem.Set(val)
			} else {
				return fmt.Errorf("cannot set destination  value for column %s", string(r.defs[i].Name))
			}
		} else {
			scanner, ok := destVal.Interface().(interface{ Scan(interface{}) error })

			if !ok {
				return fmt.Errorf("destination kind '%v' not supported for value kind '%v' of column '%s'",
					destVal.Elem().Kind(), val.Kind(), string(r.defs[i].Name))
			}
			if err := scanner.Scan(val.Interface()); err != nil {
				return fmt.Errorf("scanning value error for column '%s': %w", string(r.defs[i].Name), err)
			}

		}
	}
	return r.nextErr
}

type MockRow struct {
	commandTag pgconn.CommandTag
	defs       []pgproto3.FieldDescription
	row        []interface{}
	nextErr    error
}

// NewMockRow generates new mock row.
func NewMockRow(columns []string) *MockRow {
	var coldefs []pgproto3.FieldDescription
	for _, column := range columns {
		coldefs = append(coldefs, pgproto3.FieldDescription{Name: []byte(column)})
	}
	return &MockRow{
		defs:    coldefs,
		nextErr: nil,
	}
}

func (r *MockRow) RowError(row int, err error) *MockRow {
	r.nextErr = err
	return r
}

func (r *MockRow) AddRow(values ...interface{}) *MockRow {
	if len(values) != len(r.defs) {
		panic("expected number of values to match number of columns")
	}
	row := make([]interface{}, len(r.defs))
	copy(row, values)
	r.row = row
	return r
}

func (r *MockRow) AddCommandTag(tag pgconn.CommandTag) *MockRow {
	r.commandTag = tag
	return r
}

func (r *MockRow) FromCSVString(s string) *MockRow {
	res := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(res)
	for {
		res, err := csvReader.Read()
		if err != nil || res == nil {
			break
		}
		row := make([]interface{}, len(r.defs))
		for i, v := range res {
			row[i] = CSVColumnParser(strings.TrimSpace(v))
		}
		r.row = row
	}
	return r
}

type mockRowSetWithDefinition struct {
	*mockRowSet
}

// NewMockRowWithColumnDefinition generates mock row with columns metadata
func NewMockRowWithColumnDefinition(columns ...pgproto3.FieldDescription) *MockRow {
	return &MockRow{
		defs:    columns,
		nextErr: nil,
	}
}

func (r *MockRow) Compose() pgx.Row {
	return composeRow(r)
}

func composeRow(row *MockRow) pgx.Row {
	var pgxrow pgx.Row
	if row.defs != nil {
		pgxrow = &mockRowSetWithDefinition{&mockRowSet{set: row}}
	} else {
		pgxrow = &mockRowSet{set: row}
	}
	return pgxrow
}
