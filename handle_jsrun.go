package taproot

import (
	"bytes"
	"context"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/constants"
	"github.com/highgrav/taproot/v1/jsrun"
	"github.com/highgrav/taproot/v1/logging"
	"github.com/julienschmidt/httprouter"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"io"
	"net/http"
	"strings"
)

// Injects data about an HTTP request into a JS runtime
func injectHttpRequest(r *http.Request, vm *goja.Runtime) {
	// First, check to see if there's a correlation ID in context
	corrID := ""
	if r.Context().Value(constants.HTTP_CONTEXT_CORRELATION_KEY) != nil {
		corrID = r.Context().Value(constants.HTTP_CONTEXT_CORRELATION_KEY).(string)
	}

	// Check to see if there's a content security policy nonce
	cspNonce := ""
	if r.Context().Value(constants.HTTP_CONTEXT_CSP_NONCE_KEY) != nil {
		corrID = r.Context().Value(constants.HTTP_CONTEXT_CSP_NONCE_KEY).(string)
	}

	sessionKey := ""
	if r.Context().Value(constants.HTTP_CONTEXT_SESSION_KEY) != nil {
		sessionKey = r.Context().Value(constants.HTTP_CONTEXT_SESSION_KEY).(string)
	}

	userKey := ""
	if r.Context().Value(constants.HTTP_CONTEXT_USER_KEY) != nil {
		user, ok := r.Context().Value(constants.HTTP_CONTEXT_USER_KEY).(authn.User)
		if ok {
			userKey = user.UserID
		}
	}

	type requestData struct {
		Host        string              `json:"host"`
		Method      string              `json:"method"`
		URL         string              `json:"url"`
		Params      map[string]string   `json:"params"`
		QueryString string              `json:"query"`
		Body        string              `json:"body"`
		Form        map[string][]string `json:"form"`
	}
	reqData := requestData{}
	reqData.Params = make(map[string]string)
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

	// get URL paramters
	ps := httprouter.ParamsFromContext(r.Context())
	for _, p := range ps {
		reqData.Params[p.Key] = p.Value
	}

	// Get the body
	bodyData, err := io.ReadAll(r.Body)
	if err == nil {
		reqData.Body = string(bodyData)
	}
	// reset the body so we're polite to the next middleware that needs to read it (there shouldn't be any!)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyData))

	flaglist := make(map[string]any)
	if userKey != "" || sessionKey != "" {
		var flaguser ffuser.User
		if userKey != "" {
			flaguser = ffuser.NewUser(userKey)
		} else if sessionKey != "" {
			flaguser = ffuser.NewUser(sessionKey)
		}
		allFlags := ffclient.AllFlagsState(flaguser)
		for k, v := range allFlags.GetFlags() {
			flaglist[k] = v
		}
	}

	vm.Set("correlationId", corrID)
	vm.Set("cspNonce", cspNonce)
	vm.Set("sessionId", sessionKey)
	vm.Set("userId", userKey)
	vm.Set("flags", flaglist)
	vm.Set("request", reqData)
}

// Injects some utility functions into the JS runtime
func addJSUtilFunctor(svr *AppServer, vm *goja.Runtime) {
	obj := vm.NewObject()

	printToStdout := func(val goja.Value) {
		logging.LogToDeck(context.Background(), "info", "JS", "output", val.String())
	}

	obj.Set("print", printToStdout)
	vm.Set("util", obj)
}

// An endpoint route that executes a compiled script identified by the path to the script, injecting various data and functions into the runtime.
func (srv *AppServer) HandleScript(scriptKey string, cachedDuration int, customCtx *map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// special case for when we have .jsml file names
		if strings.HasSuffix(scriptKey, ".jsml") {
			scriptKey = scriptKey[:len(scriptKey)-2]
		}

		if cachedDuration > 0 {
			// get cached version
			// Note that we are caching based on the JS name, not the path
			cachedStr, ok := srv.PageCache.Get(scriptKey)
			if ok {
				w.WriteHeader(200)
				w.Write([]byte(cachedStr))
				return
			}
		}

		script, err := srv.js.GetScript(scriptKey)
		if err != nil {
			logging.LogToDeck(r.Context(), "info", "JS", "error", err.Error())
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
		if ctx.Value(constants.HTTP_CONTEXT_USER_KEY) != nil {
			ctxItems["user"] = ctx.Value(constants.HTTP_CONTEXT_USER_KEY)
		} else {
			ctxItems["user"] = authn.User{}
		}
		if ctx.Value(constants.HTTP_CONTEXT_REALM_KEY) != nil {
			ctxItems["realm"] = ctx.Value(constants.HTTP_CONTEXT_REALM_KEY)
		} else {
			ctxItems["realm"] = ""
		}
		if ctx.Value(constants.HTTP_CONTEXT_DOMAIN_KEY) != nil {
			ctxItems["domain"] = ctx.Value(constants.HTTP_CONTEXT_DOMAIN_KEY)
		} else {
			ctxItems["domain"] = ""
		}
		if ctx.Value(constants.HTTP_CONTEXT_ACACIA_RIGHTS_KEY) != nil {
			ctxItems["rights"] = ctx.Value(constants.HTTP_CONTEXT_ACACIA_RIGHTS_KEY)
		} else {
			ctxItems["rights"] = []string{}
		}
		if ctx.Value(constants.HTTP_CONTEXT_CORRELATION_KEY) != nil {
			ctxItems["correlationId"] = ctx.Value(constants.HTTP_CONTEXT_CORRELATION_KEY)
		} else {
			ctxItems["correlationId"] = ""
		}
		if ctx.Value(constants.HTTP_CONTEXT_CSP_NONCE_KEY) != nil {
			ctxItems["cspNonceId"] = ctx.Value(constants.HTTP_CONTEXT_CSP_NONCE_KEY)
		} else {
			ctxItems["cspNonceId"] = ""
		}
		jsrun.InjectContextDataFunctor(ctxItems, "context", vm)

		// Pass in any custom data
		if customCtx != nil {
			jsrun.InjectContextDataFunctor(*customCtx, "data", vm)
		}

		bufwriter := common.NewBufferedHttpResponseWriter(w)

		jsrun.InjectJSHttpFunctor(w, r, bufwriter, vm)
		jsrun.InjectJSDBFunctor(srv.DBs, vm)
		addJSUtilFunctor(srv, vm)

		for _, v := range srv.jsinjections {
			v(vm)
		}

		logging.LogToDeck(r.Context(), "info", "JS", "run", "running "+scriptKey)
		injectHttpRequest(r, vm)
		_, err = vm.RunProgram(script)

		if jserr, ok := err.(*goja.Exception); ok {
			logging.LogToDeck(r.Context(), "error", "JS", "fail", "error running "+scriptKey+": "+jserr.Error())
			srv.ErrorResponse(w, r, http.StatusInternalServerError, jserr.Error())
		} else if err != nil {
			logging.LogToDeck(r.Context(), "error", "JS", "fail", "error running "+scriptKey+": "+err.Error())
			srv.ErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		} else {
			logging.LogToDeck(r.Context(), "info", "JS", "done", "completed "+scriptKey)
		}

		// flush bufwriter
		if !bufwriter.IsClosed {
			bufwriter.Flush()
		}

		if cachedDuration > 0 {
			// cache result
			if bufwriter.Code == 200 {
				srv.PageCache.Put(scriptKey, bufwriter.Result.String(), cachedDuration)
			}
		}

	}
}
