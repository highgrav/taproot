package taproot

import (
	"bytes"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/google/deck"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/jsrun"
	"io"
	"net/http"
)

func injectHttpRequest(r *http.Request, vm *goja.Runtime) {
	// First, check to see if there's a correlation ID in context
	corrID := ""
	if r.Context().Value(HTTP_CONTEXT_CORRELATION_KEY) != nil {
		corrID = r.Context().Value(HTTP_CONTEXT_CORRELATION_KEY).(string)
	}

	// Check to see if there's a content security policy nonce
	cspNonce := ""
	if r.Context().Value(HTTP_CONTEXT_CSP_NONCE_KEY) != nil {
		corrID = r.Context().Value(HTTP_CONTEXT_CSP_NONCE_KEY).(string)
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
func (srv *AppServer) HandleScript(scriptKey string, customCtx *map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var corrId string = r.Context().Value(HTTP_CONTEXT_CORRELATION_KEY).(string)

		script, err := srv.js.GetScript(scriptKey)
		if err != nil {
			deck.Error("JS\t" + corrId + "\t" + err.Error())
			w.WriteHeader(http.StatusExpectationFailed)
			return
		}
		vm := goja.New()
		vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		new(require.Registry).Enable(vm)
		console.Enable(vm)

		// Pass in the context
		ctx := r.Context()
		ctxItems := make(map[string]any)
		if ctx.Value(HTTP_CONTEXT_USER_KEY) != nil {
			ctxItems["user"] = ctx.Value(HTTP_CONTEXT_USER_KEY)
		} else {
			ctxItems["user"] = authn.User{}
		}
		if ctx.Value(HTTP_CONTEXT_REALM_KEY) != nil {
			ctxItems["realm"] = ctx.Value(HTTP_CONTEXT_REALM_KEY)
		} else {
			ctxItems["realm"] = ""
		}
		if ctx.Value(HTTP_CONTEXT_DOMAIN_KEY) != nil {
			ctxItems["domain"] = ctx.Value(HTTP_CONTEXT_DOMAIN_KEY)
		} else {
			ctxItems["domain"] = ""
		}
		if ctx.Value(HTTP_CONTEXT_ACACIA_RIGHTS_KEY) != nil {
			ctxItems["rights"] = ctx.Value(HTTP_CONTEXT_ACACIA_RIGHTS_KEY)
		} else {
			ctxItems["rights"] = []string{}
		}
		if ctx.Value(HTTP_CONTEXT_CORRELATION_KEY) != nil {
			ctxItems["correlationId"] = ctx.Value(HTTP_CONTEXT_CORRELATION_KEY)
		} else {
			ctxItems["correlationId"] = ""
		}
		if ctx.Value(HTTP_CONTEXT_CSP_NONCE_KEY) != nil {
			ctxItems["cspNonceId"] = ctx.Value(HTTP_CONTEXT_CSP_NONCE_KEY)
		} else {
			ctxItems["cspNonceId"] = ""
		}
		jsrun.InjectContextDataFunctor(ctxItems, "context", vm)

		// Pass in any custom data
		if customCtx != nil {
			jsrun.InjectContextDataFunctor(*customCtx, "data", vm)
		}

		jsrun.InjectJSHttpFunctor(w, r, vm)
		jsrun.InjectJSDBFunctor(srv.DBs, vm)
		addJSUtilFunctor(srv, vm)

		for _, v := range srv.jsinjections {
			v(vm)
		}

		deck.Info(fmt.Sprintf("JS\t%s\t%s\n", corrId, scriptKey))
		injectHttpRequest(r, vm)
		_, err = vm.RunProgram(script)

		if jserr, ok := err.(*goja.Exception); ok {
			deck.Error("JS\t" + corrId + "\tError running " + scriptKey + ": " + jserr.Error())
			srv.ErrorResponse(w, r, http.StatusInternalServerError, jserr.Error())
		} else if err != nil {
			deck.Error("JS\t" + corrId + "\tError running " + scriptKey + ": " + err.Error())
			srv.ErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}

func addJSUtilFunctor(svr *AppServer, vm *goja.Runtime) {
	obj := vm.NewObject()

	printToStdout := func(val goja.Value) {
		deck.Info("%s\n", val.String())
	}

	obj.Set("print", printToStdout)
	vm.Set("util", obj)
}
