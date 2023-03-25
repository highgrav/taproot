package jsrun

import (
	"github.com/dop251/goja"
	"github.com/highgrav/taproot/v1/common"
	"net/http"
)

func InjectJSHttpFunctor(w http.ResponseWriter, r *http.Request, bufwriter *common.BufferedHttpResponseWriter, vm *goja.Runtime) {
	obj := vm.NewObject()

	writeToHttp := func(val goja.Value) {
		exp := val.String()

		_, err := bufwriter.Write([]byte(exp))
		if err != nil {
			// TODO
		}
	}

	writeRespCode := func(val goja.Value) {
		// TODO -- handle error
		bufwriter.Code = int(val.ToInteger())
		w.WriteHeader(int(val.ToInteger()))
	}

	flush := func() {
		bufwriter.Flush()
	}

	redirect := func(val goja.Value) {
		http.Redirect(w, r, val.String(), http.StatusSeeOther)
	}

	obj.Set("write", writeToHttp)
	obj.Set("flush", flush)
	obj.Set("redirect", redirect)
	obj.Set("responseCode", writeRespCode)
	obj.Set("isLoaded", true)
	vm.Set("http", obj)

	// The default output should always be accessible via out.write()
	outObj := vm.NewObject()
	outObj.Set("write", writeToHttp)
	vm.Set("out", outObj)
}
