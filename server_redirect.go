package taproot

func NewHttpRedirectServer() *WebServer {
	return NewWebServer(nil, HttpConfig{})
}
