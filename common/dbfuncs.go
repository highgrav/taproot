package common

import (
	"database/sql"
	"reflect"
	"strings"
)

type DBRowset struct {
	Rows []map[string]interface{}
	Cols map[string]reflect.Kind
}

func (rows *DBRowset) NewCursor() *DBRowCursor {
	c := &DBRowCursor{
		rowset:     rows,
		currentRow: 0,
		rowSize:    len(rows.Cols),
	}
	return c
}

type DBRowCursor struct {
	rowset     *DBRowset
	currentRow int
	rowSize    int
}

func (c *DBRowCursor) Next() bool {
	c.currentRow++
	if c.currentRow < c.rowSize {
		return true
	}
	return false
}

func (c *DBRowCursor) Get(col string) interface{} {

	return nil
}

func (c *DBRowCursor) TypeOf(col string) reflect.Type {
	return reflect.TypeOf(c.Get(col))
}

func (c *DBRowCursor) KindOf(col string) reflect.Kind {
	return reflect.TypeOf(c.Get(col)).Kind()
}

func RowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	var rowset []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i, _ := range values {
			ptrs[i] = &values[i]
		}
		err := rows.Scan(ptrs...)
		if err != nil {
			return nil, err
		}
		results := make(map[string]interface{})
		for i, val := range values {
			results[strings.ToLower(cols[i])] = val
		}
		rowset = append(rowset, results)
	}
	return rowset, nil
}
