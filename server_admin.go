package taproot

func (srv *AppServer) NewAdminServer(cfg HttpConfig) *WebServer {
	return NewWebServer(nil, cfg)
}
