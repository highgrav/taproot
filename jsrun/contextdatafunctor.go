package jsrun

import (
	"github.com/dop251/goja"
)

func InjectContextDataFunctor(cd map[string]any, vm *goja.Runtime) {
	obj := vm.NewObject()
	for k, v := range cd {
		err := obj.Set(k, v)
		if err != nil {
			panic(err)
		}
	}
	vm.Set("data", obj)
}
