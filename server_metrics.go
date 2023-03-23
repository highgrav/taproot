package taproot

import (
	"expvar"
	"github.com/google/deck"
	"github.com/highgrav/taproot/v1/common"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"
)

// Creates a new metrics server using an HttpConfig. Metrics servers should be filtered to only allow local connections.
func (srv *AppServer) NewMetricsServer(cfg HttpConfig, usePprof bool) *WebServer {
	ws := NewWebServer(nil, cfg)
	ws.Router.HandlerFunc(http.MethodGet, "/global", srv.metrics_handle_global)
	ws.Router.HandlerFunc(http.MethodGet, "/stats", srv.metrics_handle_path)
	ws.Router.HandlerFunc(http.MethodGet, "/", srv.metrics_handle_getpaths)

	if usePprof {
		ws.Router.HandlerFunc(http.MethodGet, "/debug/pprof/", pprof.Index)
		ws.Router.HandlerFunc(http.MethodGet, "/debug/pprof/cmdline", pprof.Cmdline)
		ws.Router.HandlerFunc(http.MethodGet, "/debug/pprof/profile", pprof.Profile)
		ws.Router.HandlerFunc(http.MethodGet, "/debug/pprof/symbol", pprof.Symbol)
		ws.Router.HandlerFunc(http.MethodGet, "/debug/pprof/trace", pprof.Trace)
		ws.Router.Handler(http.MethodGet, "/debug/pprof/goroutine", pprof.Handler("goroutine"))
		ws.Router.Handler(http.MethodGet, "/debug/pprof/heap", pprof.Handler("heap"))
		ws.Router.Handler(http.MethodGet, "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		ws.Router.Handler(http.MethodGet, "/debug/pprof/block", pprof.Handler("block"))

		ws.Router.HandlerFunc(http.MethodGet, "/gc/force", func(w http.ResponseWriter, r *http.Request) {
			runtime.GC()
			de := DataEnvelope{}
			de["ok"] = true
			srv.WriteJSON(w, true, 200, de, nil)
		})
	}

	return ws
}

// Dumps global stats
func (srv *AppServer) metrics_handle_global(w http.ResponseWriter, r *http.Request) {
	st := srv.globalStats

	st2 := make(map[string]any)
	st2["requests"] = st.requests.Value()
	st2["responses"] = st.responses.Value()
	st2["processing_time"] = st.processingTime.Value()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	st2["mem__mem_alloc_mb"] = common.BToMb(m.Alloc)
	st2["mem__total_mem_alloc_mb"] = common.BToMb(m.TotalAlloc)
	st2["mem__os_mem_used_mb"] = common.BToMb(m.Sys)
	st2["gc__total_gcs"] = m.NumGC
	st2["mem__mallocs"] = m.Mallocs
	st2["mem__frees"] = m.Frees
	st2["mem__heap_alloc_mb"] = common.BToMb(m.HeapAlloc)
	st2["mem__heap_in_use_mb"] = common.BToMb(m.HeapInuse)
	st2["mem__heap_released_mb"] = common.BToMb(m.HeapReleased)
	st2["gc__heap_objects"] = m.HeapObjects
	st2["mem__stack_mb"] = common.BToMb(m.StackSys)
	st2["gc_stop_the_world_ns"] = m.PauseTotalNs
	st2["uptime_secs"] = time.Now().Sub(srv.startedOn).Seconds()

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

	hist := st.window.MakeHistogram()
	hist2 := hist.Xile(20)
	st2["histogram"] = hist2

	env := DataEnvelope{}
	env["ok"] = true
	env["stats"] = st2
	err := srv.WriteJSON(w, true, 200, env, nil)
	if err != nil {
		deck.Error("metrics server path '" + path + "' stats: " + err.Error())
	}
}
