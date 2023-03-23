package jsrun

import (
	"github.com/dop251/goja"
	"net/http"
)

func InjectJSFFlagFunctor(r *http.Request, vm *goja.Runtime) {
	obj := vm.NewObject()

	getFlags := func() {

	}

	obj.Set("getFlags", getFlags)
	obj.Set("isLoaded", true)
	vm.Set("fflags", obj)
}
