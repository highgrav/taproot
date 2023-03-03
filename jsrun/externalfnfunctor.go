package jsrun

import (
	"github.com/dop251/goja"
)

type ExternalJSFunction func(...any) *JSCallReturnValue

func InjectJSFnFunctor(externalFns *map[string]ExternalJSFunction, vm *goja.Runtime) {
	obj := vm.NewObject()

	executeFn := func(args ...goja.Value) *JSCallReturnValue {
		retval := &JSCallReturnValue{}
		if len(args) < 1 {
			retval.OK = false
			retval.ResultCode = -9123
			retval.ResultDescription = "Error (see errors array)"
			retval.Errors = []string{"First argument must be an importable function name"}
			retval.Results = make(map[string]interface{})
			return retval
		}
		fnName := args[0].String()
		fn, ok := (*externalFns)[fnName]
		if !ok {
			retval.OK = false
			retval.ResultCode = -9123
			retval.ResultDescription = "Error (see errors array)"
			retval.Errors = []string{"Unknown importable argument name '" + fnName + "'"}
			retval.Results = make(map[string]interface{})
			return retval
		}
		return fn(args[1:])

		return retval
	}

	obj.Set("exec", executeFn)
	obj.Set("isLoaded", true)
	vm.Set("fns", obj)
}
