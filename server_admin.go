package taproot

import "net/http"

func (srv *AppServer) NewAdminServer(cfg HttpConfig) *WebServer {
	return NewWebServer(nil, cfg)
}

func (srv *AppServer) admin_handle_script_cache(w http.ResponseWriter, r *http.Request) {

}

func (srv *AppServer) admin_handle_pause(w http.ResponseWriter, r *http.Request) {

}

func (srv *AppServer) admin_handle_shutdown(w http.ResponseWriter, r *http.Request) {

}

func (srv *AppServer) admin_handle_ip_filter(w http.ResponseWriter, r *http.Request) {

}

func (srv *AppServer) admin_handle_rate_limit(w http.ResponseWriter, r *http.Request) {

}

func (srv *AppServer) admin_handle_acacia(w http.ResponseWriter, r *http.Request) {

}

func (srv *AppServer) admin_handle_acacia_add(w http.ResponseWriter, r *http.Request) {

}

func (srv *AppServer) admin_handle_acacia_flush(w http.ResponseWriter, r *http.Request) {

}
