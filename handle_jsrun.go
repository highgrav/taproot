package taproot

import (
	"bytes"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/google/deck"
	"highgrav/taproot/v1/jsrun"
	"io"
	"net/http"
)

func injectHttpRequest(r *http.Request, vm *goja.Runtime) {
	// First, check to see if there's a correlation ID in context
	corrID := ""
	if r.Context().Value(CONTEXT_CORRELATION_KEY_NAME) != nil {
		corrID = r.Context().Value(CONTEXT_CORRELATION_KEY_NAME).(string)
	}

	// Check to see if there's a content security policy nonce
	cspNonce := ""
	if r.Context().Value(CONTEXT_CSP_NONCE_KEY_NAME) != nil {
		corrID = r.Context().Value(CONTEXT_CSP_NONCE_KEY_NAME).(string)
	}
	type requestData struct {
		Host        string              `json:"host"`
		Method      string              `json:"method"`
		URL         string              `json:"url"`
		QueryString string              `json:"query"`
		Body        string              `json:"body"`
		Form        map[string][]string `json:"form"`
	}
	reqData := requestData{}
	r.ParseForm()
	reqData.Host = r.Host
	reqData.Method = r.Method
	reqData.URL = r.URL.String()
	reqData.QueryString = r.URL.RawQuery
	formElems := make(map[string][]string)
	for key, val := range r.Form {
		formElems[key] = val
	}
	reqData.Form = formElems

	// Get the body
	bodyData, err := io.ReadAll(r.Body)
	if err == nil {
		reqData.Body = string(bodyData)
	}
	// reset the body so we're polite to the next middleware that needs to read it (there shouldn't be any!)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyData))

	vm.Set("correlationId", corrID)
	vm.Set("cspNonce", cspNonce)
	vm.Set("req", reqData)
}

// An endpoint route that executes a compiled script
func (svr *AppServer) HandleScript(scriptKey string, ctx *map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		script, err := svr.js.GetScript(scriptKey)
		if err != nil {
			w.WriteHeader(http.StatusExpectationFailed)
			return
		}
		vm := goja.New()
		vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		new(require.Registry).Enable(vm)
		console.Enable(vm)
		if ctx != nil {
			jsrun.InjectContextDataFunctor(*ctx, vm)
		}
		jsrun.InjectJSHttpFunctor(w, r, vm)
		jsrun.InjectJSDBFunctor(svr.DBs, vm)
		addJSUtilFunctor(svr, vm)

		for _, v := range svr.jsinjections {
			v(vm)
		}

		corrId := r.Context().Value(CONTEXT_CORRELATION_KEY_NAME)
		deck.Info(fmt.Sprintf("JS\t%s\t%s\n", corrId, scriptKey))
		injectHttpRequest(r, vm)
		_, err = vm.RunProgram(script)

		if jserr, ok := err.(*goja.Exception); ok {
			deck.Error("Error running " + scriptKey + ": " + jserr.Error())
			svr.ErrorResponse(w, r, http.StatusInternalServerError, jserr.Error())
		} else if err != nil {
			deck.Error("Error running " + scriptKey + ": " + err.Error())
			svr.ErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}

func addJSUtilFunctor(svr *AppServer, vm *goja.Runtime) {
	obj := vm.NewObject()

	printToStdout := func(val goja.Value) {
		fmt.Printf("%s\n", val.String())
	}

	obj.Set("print", printToStdout)
	vm.Set("util", obj)
}
