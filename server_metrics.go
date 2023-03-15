package taproot

func (srv *AppServer) NewMetricsServer() *WebServer {
	return NewWebServer(nil, HttpConfig{})
}
