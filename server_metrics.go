package taproot

func NewMetricsServer() *WebServer {
	return NewWebServer(nil, HttpConfig{})
}
