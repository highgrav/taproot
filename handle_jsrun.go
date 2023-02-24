package taproot

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"highgrav/taproot/v1/jsrun"
	"net/http"
)

// An endpoint route that executes a compiled script
func (svr *Server) HandleScript(scriptKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		script, err := svr.js.GetScript(scriptKey)
		if err != nil {
			w.WriteHeader(http.StatusExpectationFailed)
			return
		}
		vm := goja.New()
		new(require.Registry).Enable(vm)
		console.Enable(vm)
		jsrun.InjectJSHttpFunctor(w, r, vm)
		jsrun.InjectJSDBFunctor(svr.DBs, vm)
		addJSUtilFunctor(svr, vm)

		for _, v := range svr.jsinjections {
			v(vm)
		}

		// TODO -- inject user info

		_, err = vm.RunProgram(script)
		if err != nil {
			svr.ErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}

func addJSUtilFunctor(svr *Server, vm *goja.Runtime) {
	obj := vm.NewObject()

	printToStdout := func(val goja.Value) {
		fmt.Printf("%s\n", val.String())
	}

	obj.Set("print", printToStdout)
	vm.Set("util", obj)
}
