package jsrun

import (
	"github.com/dop251/goja"
	"net/http"
)

func InjectJSHttpFunctor(w http.ResponseWriter, r *http.Request, vm *goja.Runtime) {
	obj := vm.NewObject()

	writeToHttp := func(val goja.Value) {
		exp := val.String()

		_, err := w.Write([]byte(exp))
		if err != nil {
			// TODO
		}
	}

	writeRespCode := func(val goja.Value) {
		w.WriteHeader(int(val.ToInteger()))
	}

	obj.Set("write", writeToHttp)
	obj.Set("responseCode", writeRespCode)
	obj.Set("isLoaded", true)
	vm.Set("http", obj)
	// TODO -- We should have an "out.write()" alias that defaults to http.write() or whatever the preferred output
	// TODO -- mode is (e.g., filesystem, pdf, etc.)
}
