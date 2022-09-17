package pgxpoolgo

import (
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"reflect"
)

/*
The codes below is based on `pgxmock` from `github.com/pashagolub/pgxmock`
*/

type MockRow struct {
	commandTag pgconn.CommandTag
	defs       []pgproto3.FieldDescription
	row        []interface{}
	scanErr    error
}

type row struct {
	row *MockRow
}

// NewMockRow mocks pgx.Row.
func NewMockRow(columns []string) *MockRow {
	var coldefs []pgproto3.FieldDescription
	for _, column := range columns {
		coldefs = append(coldefs, pgproto3.FieldDescription{Name: []byte(column)})
	}
	return &MockRow{
		defs:    coldefs,
		scanErr: nil,
	}
}

func (mr *MockRow) ScanError(err error) *MockRow {
	mr.scanErr = err
	return mr
}

func (mr *MockRow) AddRow(values ...interface{}) *MockRow {
	if len(values) != len(mr.defs) {
		panic("expected number of values to match number of columns")
	}
	newRow := make([]interface{}, len(mr.defs))
	copy(newRow, values)
	mr.row = newRow
	return mr
}

func (mr *MockRow) Compose() pgx.Row {
	return &row{row: mr}
}

func (r *row) Scan(dest ...interface{}) error {
	currentRow := r.row
	if currentRow.scanErr != nil {
		return currentRow.scanErr
	}
	if len(dest) != len(currentRow.defs) {
		return fmt.Errorf("incorrect argument number %d for columns %d", len(dest), len(currentRow.defs))
	}
	for i, col := range currentRow.row {
		if dest[i] == nil {
			continue
		}
		destVal := reflect.ValueOf(dest[i])
		if destVal.Kind() != reflect.Ptr {
			return fmt.Errorf("destination argument must be a pointer for column %s", currentRow.defs[i].Name)
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
				return fmt.Errorf("cannot set destination  value for column %s", string(currentRow.defs[i].Name))
			}
		} else {
			scanner, ok := destVal.Interface().(interface{ Scan(interface{}) error })
			if !ok {
				return fmt.Errorf("destination kind '%v' not supported for value kind '%v' of column '%s'",
					destVal.Elem().Kind(), val.Kind(), string(currentRow.defs[i].Name))
			}
			if err := scanner.Scan(val.Interface()); err != nil {
				return fmt.Errorf("scanning value error for column '%s': %w", string(currentRow.defs[i].Name), err)
			}
		}
	}
	return nil
}
