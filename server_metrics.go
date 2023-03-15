package taproot

import (
	"expvar"
	"github.com/google/deck"
	"net/http"
)

// Creates a new metrics server using an HttpConfig. Metrics servers should be filtered to only allow local connections.
func (srv *AppServer) NewMetricsServer(cfg HttpConfig) *WebServer {
	ws := NewWebServer(nil, cfg)
	ws.Router.HandlerFunc(http.MethodGet, "/global", srv.metrics_handle_global)
	ws.Router.HandlerFunc(http.MethodGet, "/stats", srv.metrics_handle_path)
	ws.Router.HandlerFunc(http.MethodGet, "/", srv.metrics_handle_getpaths)
	return ws
}

// Dumps global stats
func (srv *AppServer) metrics_handle_global(w http.ResponseWriter, r *http.Request) {
	st := srv.globalStats

	st2 := make(map[string]any)
	st2["requests"] = st.requests.Value()
	st2["responses"] = st.responses.Value()
	st2["processing_time"] = st.processingTime.Value()
	ss := make(map[string]string)
	st.responseCodes.Do(func(kv expvar.KeyValue) {
		ss[kv.Key] = kv.Value.String()
	})
	st2["response_codes"] = ss

	env := DataEnvelope{}
	env["ok"] = true
	env["stats"] = st2
	err := srv.WriteJSON(w, true, 200, env, nil)
	if err != nil {
		deck.Error("metrics server global stats: " + err.Error())
	}
}

// Dumps a list of all paths registered for stats
func (srv *AppServer) metrics_handle_getpaths(w http.ResponseWriter, r *http.Request) {
	paths := make([]string, 0)
	for k, _ := range srv.stats {
		paths = append(paths, k)
	}
	env := DataEnvelope{}
	env["ok"] = true
	env["paths"] = paths
	err := srv.WriteJSON(w, true, 200, env, nil)
	if err != nil {
		deck.Error("metrics server get paths: " + err.Error())
	}
}

// Returns stats for a specific path
func (srv *AppServer) metrics_handle_path(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	if !vals.Has("path") {
		srv.ErrorResponse(w, r, 400, "path query parameter not included")
		return
	}
	path := vals.Get("path")
	if path == "" {
		srv.ErrorResponse(w, r, 400, "path query parameter blank or empty")
		return
	}

	st, ok := srv.stats[path]
	if !ok {
		srv.ErrorResponse(w, r, 410, "path '"+path+"' does not exist")
		return
	}
	st2 := make(map[string]any)
	st2["path"] = path
	st2["requests"] = st.requests.Value()
	st2["responses"] = st.responses.Value()
	st2["processing_time"] = st.processingTime.Value()
	ss := make(map[string]string)
	st.responseCodes.Do(func(kv expvar.KeyValue) {
		ss[kv.Key] = kv.Value.String()
	})
	st2["response_codes"] = ss

	env := DataEnvelope{}
	env["ok"] = true
	env["stats"] = st2
	err := srv.WriteJSON(w, true, 200, env, nil)
	if err != nil {
		deck.Error("metrics server path '" + path + "' stats: " + err.Error())
	}
}
