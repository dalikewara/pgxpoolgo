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

type MockRows struct {
	commandTag pgconn.CommandTag
	defs       []pgproto3.FieldDescription
	rows       [][]interface{}
	index      int
	scanErr    map[int]error
}

type rows struct {
	rows  []*MockRows
	index int
}

// NewMockRows mocks pgx.Rows.
func NewMockRows(columns []string) *MockRows {
	var coldefs []pgproto3.FieldDescription
	for _, column := range columns {
		coldefs = append(coldefs, pgproto3.FieldDescription{Name: []byte(column)})
	}
	return &MockRows{
		defs:    coldefs,
		scanErr: make(map[int]error),
	}
}

func (mr *MockRows) ScanError(rowIndex int, err error) *MockRows {
	mr.scanErr[rowIndex] = err
	return mr
}

func (mr *MockRows) AddRow(values ...interface{}) *MockRows {
	if len(values) != len(mr.defs) {
		panic("expected number of values to match number of columns")
	}
	newRow := make([]interface{}, len(mr.defs))
	copy(newRow, values)
	mr.rows = append(mr.rows, newRow)
	return mr
}

func (mr *MockRows) AddCommandTag(tag pgconn.CommandTag) *MockRows {
	mr.commandTag = tag
	return mr
}

func (mr *MockRows) Compose() pgx.Rows {
	return &rows{rows: []*MockRows{mr}}
}

func (r *rows) Close() {}

func (r *rows) Err() error {
	currentRow := r.rows[r.index]
	return currentRow.scanErr[currentRow.index-1]
}

func (r *rows) CommandTag() pgconn.CommandTag {
	return r.rows[r.index].commandTag
}

func (r *rows) FieldDescriptions() []pgproto3.FieldDescription {
	return r.rows[r.index].defs
}

func (r *rows) Next() bool {
	currentRow := r.rows[r.index]
	currentRow.index++
	return currentRow.index <= len(currentRow.rows)
}

func (r *rows) Scan(dest ...interface{}) error {
	currentRow := r.rows[r.index]
	if currentRow.scanErr[currentRow.index-1] != nil {
		return currentRow.scanErr[currentRow.index-1]
	}
	if len(dest) != len(currentRow.defs) {
		return fmt.Errorf("incorrect argument number %d for columns %d", len(dest), len(currentRow.defs))
	}
	for i, col := range currentRow.rows[currentRow.index-1] {
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

func (r *rows) Values() ([]interface{}, error) {
	currentRow := r.rows[r.index]
	return currentRow.rows[currentRow.index-1], currentRow.scanErr[currentRow.index-1]
}

func (r *rows) RawValues() [][]byte {
	currentRow := r.rows[r.index]
	dest := make([][]byte, len(currentRow.defs))
	for i, col := range currentRow.rows[currentRow.index-1] {
		if b, ok := rawBytes(col); ok {
			dest[i] = b
			continue
		}
		dest[i] = col.([]byte)
	}
	return dest
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
