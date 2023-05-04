package jsrun

import (
	"github.com/dop251/goja"
)

const JS_EXPECTED_INTERRUPT string = "JS_EXPECTED_INTERRUPT"

func InjectJSSysFunctor(vm *goja.Runtime) {
	obj := vm.NewObject()
	exit := func() {
		vm.Interrupt(JS_EXPECTED_INTERRUPT)
	}
	obj.Set("loaded", true)
	obj.Set("exit", exit)
	vm.Set("system", obj)
}
