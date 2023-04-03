package jsrun

import (
	"context"
	"github.com/dop251/goja"
)

type InjectorFunc func(ctx context.Context, vm *goja.Runtime)
