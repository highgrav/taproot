package taproot

func (srv *AppServer) NewAdminServer() *WebServer {
	return NewWebServer(nil, HttpConfig{})
}
