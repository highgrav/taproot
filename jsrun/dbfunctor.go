package jsrun

import (
	"database/sql"
	"fmt"
	"github.com/dop251/goja"
	"highgrav/taproot/v1/common"
	"reflect"
)

/*
InjectJSDBFunctor() injects into a JS runtime a list of DB connections and a function to query them.

The injected functions appear as a top-level object named 'db'.

From within the JS function, call db.query(DSN_NAME, SQL_STATEMENT, ...), where the variadic args are parameterized values.

You can dump a list of values using db.print(...).
*/
func InjectJSDBFunctor(dbs map[string]*sql.DB, vm *goja.Runtime) {
	obj := vm.NewObject()

	dbQuery := func(args ...goja.Value) *JSCallReturnValue {
		retval := &JSCallReturnValue{}
		if len(args) < 2 {
			retval.OK = false
			retval.ResultCode = -9123
			retval.ResultDescription = "Error (see errors array)"
			retval.Errors = []string{"First argument must be a DSN reference, second argument must be a SQL string"}
			retval.Results = make(map[string]interface{})
			return retval
		}

		if args[0].ExportType() != reflect.TypeOf("") {
			retval.OK = false
			retval.ResultCode = -9124
			retval.ResultDescription = "Error (see errors array)"
			retval.Errors = []string{"First argument must be a DSN reference"}
			retval.Results = make(map[string]interface{})
			return retval
		}

		if args[1].ExportType() != reflect.TypeOf("") {
			retval.OK = false
			retval.ResultCode = -9125
			retval.ResultDescription = "Error (see errors array)"
			retval.Errors = []string{"Second argument must be a SQL query"}
			retval.Results = make(map[string]interface{})
			return retval
		}

		db, ok := dbs[args[0].String()]
		if !ok {
			retval.OK = false
			retval.ResultCode = -9126
			retval.ResultDescription = "Error (see errors array)"
			retval.Errors = []string{"DSN " + args[0].String() + " is not a valid DSN reference"}
			retval.Results = make(map[string]interface{})
			return retval
		}

		var stmt = args[0].String()
		var sqlArgs []any = make([]any, 0)
		var sqlJsArgs []goja.Value
		if len(args) > 1 {
			sqlJsArgs = args[2:]
		}
		// We're only going to support a subset of possible value types here
		var errorsList []string
		for _, v := range sqlJsArgs {
			sqlArgs = append(sqlArgs, v.String())
		}

		if len(errorsList) > 0 {
			retval.OK = false
			retval.ResultCode = -9393
			retval.ResultDescription = "Errors (see errors array)"
			retval.Errors = errorsList
			retval.Results = make(map[string]interface{})
			return retval
		}

		rows, err := db.Query(stmt, sqlArgs...)
		defer rows.Close()
		if err != nil {
			retval.OK = false
			retval.ResultCode = -9393
			retval.ResultDescription = "Errors (see errors array)"
			retval.Errors = []string{err.Error()}
			retval.Results = make(map[string]interface{})
			return retval
		}

		rowset, err := common.RowsToMap(rows)
		if err != nil {
			retval.OK = false
			retval.ResultCode = -9642
			retval.ResultDescription = "Errors (see errors array)"
			retval.Errors = []string{err.Error()}
			retval.Results = make(map[string]interface{})
			return retval
		}

		retval.OK = true
		retval.ResultCode = 200
		retval.ResultDescription = "OK"
		retval.Errors = []string{}
		retval.Results = make(map[string]interface{})
		retval.Results["rows"] = rowset
		return retval
	}

	fc := func(args ...goja.Value) {
		for _, v := range args {
			fmt.Println(v.String())
		}
	}

	obj.Set("query", dbQuery)
	obj.Set("print", fc)
	vm.Set("db", obj)
}
