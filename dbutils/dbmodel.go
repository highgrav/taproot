package dbutils

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/highgrav/taproot/common"
	"reflect"
	"strconv"
	"strings"
	"time"
)

/*
	The following tags are used:
		dbschema: schema name
		dbtable: table name
		dbcol: column name
*/

/*
	GenerateSelectFromModel(model) takes a struct value that conforms to certain expectations, and returns a simple SELECT SQL statement in postgres syntax.

The struct is expected to have a 'dbtable' tag on a field that names the table; an optional 'dbschema' tag that names the schema; and one or more 'dbcol'
tags that name the database column corresponding to the field. Fields that do not have a 'dbcol' tag are ignored. A field that has a 'dbcol' tag must be one
of the sql.Null* types (e.g., sql.NullString), or else an error will be thrown.

Example:

	// DBTable is a convenience empty struct, and can be omitted if the dbschema and dbtable
	// tags are listed on another field.
	type myModel struct {
			DBTable            `dbschema:"myschema" dbtable:"mytable"`
			DoNotUseThis       sql.NullString `dbcol:"do_not_use_this"`
			DoNotUseThisEither sql.NullTime   `dbcol:"do_not_use_this_either"`
			MyString           sql.NullString `dbcol:"my_string"`
			MyBool             sql.NullBool   `dbcol:"my_bool"`
			MyInt64            sql.NullInt64  `dbcol:"my_int_64"`
			MyInt32            sql.NullInt32  `dbcol:"my_int_32"`
			MyInt16            sql.NullInt16  `dbcol:"my_int_16"`
			MyTime             sql.NullTime   `dbcol:"my_time"`
		}

		m := myModel{
			DBTable:  DBTable{},
			MyString: sql.NullString{"Hello 'my' \"BABY\"", true},
			MyBool:   sql.NullBool{true, true},
			MyInt64:  sql.NullInt64{64, true},
			MyInt32:  sql.NullInt32{32, true},
			MyInt16:  sql.NullInt16{16, true},
			MyTime:   sql.NullTime{time.Now(), true},
		}
		sql, err := GenerateSelectFromModel(m)
		// SELECT * FROM myschema.mytable WHERE my_string = 'Hello ''my'' "BABY"' AND my_bool = true AND my_int_64 = 64 AND my_int_32 = 32 AND my_int_16 = 16 AND my_time = '2023-04-14T21:13:56-05:00'::timestamptz;
*/
func GenerateSelectFromModel(model any) (string, error) {
	schema := ""
	table := ""
	cols := make(map[string]any)
	t := reflect.TypeOf(model)

	for x := 0; x < t.NumField(); x++ {
		field := t.Field(x)
		if field.Tag.Get("dbschema") != "" {
			schema = field.Tag.Get("dbschema")
		}
		if field.Tag.Get("dbtable") != "" {
			table = field.Tag.Get("dbtable")
		}
		if field.Tag.Get("dbcol") != "" {
			fi := reflect.ValueOf(model).FieldByName(field.Name).Interface()
			switch fi := fi.(type) {
			case sql.NullString:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.String
				}
			case sql.NullBool:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Bool
				}
			case sql.NullTime:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Time
				}
			case sql.NullInt64:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Int64
				}
			case sql.NullByte:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Byte
				}
			case sql.NullFloat64:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Float64
				}
			case sql.NullInt16:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Int16
				}
			case sql.NullInt32:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Int32
				}
			default:
				return "", errors.New("attempted to use non-sql datatype on field " + field.Tag.Get("dbcol") + ": type " + fmt.Sprintf("%T", fi))
			}
		}
	}

	if table == "" {
		return "", errors.New("missing dbtable on struct")
	}

	sql := strings.Builder{}
	sql.Write([]byte("SELECT * FROM "))
	if schema != "" {
		sql.Write([]byte(schema + "."))
	}
	sql.Write([]byte(table))

	if len(cols) > 0 {
		y := 0
		sql.Write([]byte(" WHERE "))
		for k, v := range cols {
			if y > 0 {
				sql.Write([]byte(" AND "))
			}
			sql.Write([]byte(k + " = "))
			switch v.(type) {
			case string:
				sql.Write([]byte("'" + common.SanitizeStringForSql(v.(string)) + "'"))
			case bool:
				if v.(bool) == true {
					sql.Write([]byte("true"))
				} else {
					sql.Write([]byte("false"))
				}
			case time.Time:
				ts := v.(time.Time).Format("2006-01-02T15:04:05Z07:00")
				sql.Write([]byte("'" + ts + "'::timestamptz"))
			case int64:
				sql.Write([]byte(strconv.FormatInt(v.(int64), 10)))
			case int32:
				sql.Write([]byte(strconv.FormatInt(int64(v.(int32)), 10)))
			case int16:
				sql.Write([]byte(strconv.FormatInt(int64(v.(int16)), 10)))
			case float64:
				sql.Write([]byte(strconv.FormatFloat(v.(float64), 'f', -1, 64)))
			case byte:
				// TODO
			}
			y++
		}
	}

	sql.Write([]byte(";"))
	return sql.String(), nil
}

/*
GenerateInsertFromModel(model) inserts string
*/
func GenerateInsertFromModel(model any) (string, error) {
	schema := ""
	table := ""
	cols := make(map[string]any)
	t := reflect.TypeOf(model)

	for x := 0; x < t.NumField(); x++ {
		field := t.Field(x)
		if field.Tag.Get("dbschema") != "" {
			schema = field.Tag.Get("dbschema")
		}
		if field.Tag.Get("dbtable") != "" {
			table = field.Tag.Get("dbtable")
		}
		if field.Tag.Get("dbcol") != "" {
			fi := reflect.ValueOf(model).FieldByName(field.Name).Interface()
			switch fi := fi.(type) {
			case sql.NullString:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.String
				}
			case sql.NullBool:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Bool
				}
			case sql.NullTime:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Time
				}
			case sql.NullInt64:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Int64
				}
			case sql.NullByte:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Byte
				}
			case sql.NullFloat64:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Float64
				}
			case sql.NullInt16:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Int16
				}
			case sql.NullInt32:
				if fi.Valid {
					cols[field.Tag.Get("dbcol")] = fi.Int32
				}
			default:
				return "", errors.New("attempted to use non-sql datatype on field " + field.Tag.Get("dbcol") + ": type " + fmt.Sprintf("%T", fi))
			}
		}
	}

	if table == "" {
		return "", errors.New("missing dbtable on struct")
	}

	if len(cols) == 0 {
		return "", errors.New("missing dbcols on struct")
	}

	sql := strings.Builder{}
	sql.Write([]byte("INSERT INTO "))
	if schema != "" {
		sql.Write([]byte(schema + "."))
	}
	sql.Write([]byte(table))

	if len(cols) > 0 {
		colNames := make([]string, 0)
		colValues := make([]string, 0)

		for k, v := range cols {
			switch v.(type) {
			case string:
				colNames = append(colNames, k)
				colValues = append(colValues, "'"+common.SanitizeStringForSql(v.(string))+"'")
			case bool:
				colNames = append(colNames, k)
				if v.(bool) == true {
					colValues = append(colValues, "true")
				} else {
					colValues = append(colValues, "false")
				}
			case time.Time:
				colNames = append(colNames, k)
				ts := v.(time.Time).Format("2006-01-02T15:04:05Z07:00")
				colValues = append(colValues, "'"+ts+"'::timestamptz")
			case int64:
				colNames = append(colNames, k)
				colValues = append(colValues, strconv.FormatInt(v.(int64), 10))
			case int32:
				colNames = append(colNames, k)
				colValues = append(colValues, strconv.FormatInt(int64(v.(int32)), 10))
			case int16:
				colNames = append(colNames, k)
				colValues = append(colValues, strconv.FormatInt(int64(v.(int16)), 10))
			case float64:
				colNames = append(colNames, k)
				colValues = append(colValues, strconv.FormatFloat(v.(float64), 'f', -1, 64))
			case byte:
				// TODO
			}
		}
		if len(colNames) == 0 || len(colValues) == 0 {
			return "", errors.New("no column values to insert")
		}
		if len(colNames) != len(colValues) {
			return "", errors.New("unequal column name and value count")
		}
		sql.Write([]byte(" ("))
		for ii := 0; ii < len(colNames); ii++ {
			if ii > 0 {
				sql.Write([]byte(", "))
			}
			sql.Write([]byte(colNames[ii]))
		}

		sql.Write([]byte(") VALUES ("))
		for ii := 0; ii < len(colValues); ii++ {
			if ii > 0 {
				sql.Write([]byte(", "))
			}
			sql.Write([]byte(colValues[ii]))
		}
		sql.Write([]byte(")"))
	}
	sql.Write([]byte(";"))
	return sql.String(), nil
}
