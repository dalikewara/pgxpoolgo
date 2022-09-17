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

var CSVColumnParser = func(s string) interface{} {
	switch {
	case strings.ToLower(s) == "null":
		return nil
	}
	return s
}

type mockRowSets struct {
	sets []*MockRows
	pos  int
}

func (rs *mockRowSets) Err() error {
	r := rs.sets[rs.pos]
	return r.nextErr[r.pos-1]
}

func (rs *mockRowSets) CommandTag() pgconn.CommandTag {
	return rs.sets[rs.pos].commandTag
}

func (rs *mockRowSets) FieldDescriptions() []pgproto3.FieldDescription {
	return rs.sets[rs.pos].defs
}

func (rs *mockRowSets) Close() {}

func (rs *mockRowSets) Next() bool {
	r := rs.sets[rs.pos]
	r.pos++
	return r.pos <= len(r.rows)
}

func (rs *mockRowSets) Values() ([]interface{}, error) {
	r := rs.sets[rs.pos]
	return r.rows[r.pos-1], r.nextErr[r.pos-1]
}

func (rs *mockRowSets) Scan(dest ...interface{}) error {
	r := rs.sets[rs.pos]
	if len(dest) != len(r.defs) {
		return fmt.Errorf("incorrect argument number %d for columns %d", len(dest), len(r.defs))
	}
	for i, col := range r.rows[r.pos-1] {
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
	return r.nextErr[r.pos-1]
}

func (rs *mockRowSets) RawValues() [][]byte {
	r := rs.sets[rs.pos]
	dest := make([][]byte, len(r.defs))
	for i, col := range r.rows[r.pos-1] {
		if b, ok := rawBytes(col); ok {
			dest[i] = b
			continue
		}
		dest[i] = col.([]byte)
	}
	return dest
}

func (rs *mockRowSets) String() string {
	if rs.empty() {
		return "with empty rows"
	}
	msg := "should return rows:\n"
	if len(rs.sets) == 1 {
		for n, row := range rs.sets[0].rows {
			msg += fmt.Sprintf("    row %d - %+v\n", n, row)
		}
		return strings.TrimSpace(msg)
	}
	for i, set := range rs.sets {
		msg += fmt.Sprintf("    result set: %d\n", i)
		for n, row := range set.rows {
			msg += fmt.Sprintf("      row %d - %+v\n", n, row)
		}
	}
	return strings.TrimSpace(msg)
}

func (rs *mockRowSets) empty() bool {
	for _, set := range rs.sets {
		if len(set.rows) > 0 {
			return false
		}
	}
	return true
}

func rawBytes(col interface{}) (_ []byte, ok bool) {
	val, ok := col.([]byte)
	if !ok || len(val) == 0 {
		return nil, false
	}
	b := make([]byte, len(val))
	copy(b, val)
	return b, true
}

type MockRows struct {
	commandTag pgconn.CommandTag
	defs       []pgproto3.FieldDescription
	rows       [][]interface{}
	pos        int
	nextErr    map[int]error
	closeErr   error
}

// NewMockRows generates new mock rows.
func NewMockRows(columns []string) *MockRows {
	var coldefs []pgproto3.FieldDescription
	for _, column := range columns {
		coldefs = append(coldefs, pgproto3.FieldDescription{Name: []byte(column)})
	}
	return &MockRows{
		defs:    coldefs,
		nextErr: make(map[int]error),
	}
}

func (r *MockRows) CloseError(err error) *MockRows {
	r.closeErr = err
	return r
}

func (r *MockRows) RowError(row int, err error) *MockRows {
	r.nextErr[row] = err
	return r
}

func (r *MockRows) AddRow(values ...interface{}) *MockRows {
	if len(values) != len(r.defs) {
		panic("expected number of values to match number of columns")
	}
	row := make([]interface{}, len(r.defs))
	copy(row, values)
	r.rows = append(r.rows, row)
	return r
}

func (r *MockRows) AddCommandTag(tag pgconn.CommandTag) *MockRows {
	r.commandTag = tag
	return r
}

func (r *MockRows) FromCSVString(s string) *MockRows {
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
		r.rows = append(r.rows, row)
	}
	return r
}

type mockRowSetsWithDefinition struct {
	*mockRowSets
}

// NewMockRowsWithColumnDefinition generates new mock rows with columns metadata.
func NewMockRowsWithColumnDefinition(columns ...pgproto3.FieldDescription) *MockRows {
	return &MockRows{
		defs:    columns,
		nextErr: make(map[int]error),
	}
}

func (r *MockRows) Compose() pgx.Rows {
	return composeRows(r)
}

func composeRows(rows ...*MockRows) pgx.Rows {
	var pgxrows pgx.Rows
	defs := 0
	sets := make([]*MockRows, len(rows))
	for i, r := range rows {
		sets[i] = r
		if r.defs != nil {
			defs++
		}
	}
	if defs > 0 && defs == len(sets) {
		pgxrows = &mockRowSetsWithDefinition{&mockRowSets{sets: sets}}
	} else {
		pgxrows = &mockRowSets{sets: sets}
	}
	return pgxrows
}
