package taproot

func NewAdminServer() *WebServer {
	return NewWebServer(nil, HttpConfig{})
}
