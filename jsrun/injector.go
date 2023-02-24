package jsrun

import "github.com/dop251/goja"

type InjectorFunc func(vm *goja.Runtime)
