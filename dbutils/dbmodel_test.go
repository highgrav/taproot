package dbutils

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestDBModelSql(t *testing.T) {
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
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(sql)

	sql, err = GenerateInsertFromModel(m)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(sql)
}
