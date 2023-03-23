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

	redirect := func(val goja.Value) {
		http.Redirect(w, r, val.String(), http.StatusSeeOther)
	}

	obj.Set("write", writeToHttp)
	obj.Set("redirect", redirect)
	obj.Set("responseCode", writeRespCode)
	obj.Set("isLoaded", true)
	vm.Set("http", obj)

	// The default output should always be accessible via out.write()
	outObj := vm.NewObject()
	outObj.Set("write", writeToHttp)
	vm.Set("out", outObj)
}
