package taproot

func (srv *AppServer) StripHtmlInput(val string) string {
	return srv.sanitizer.StripHTML.Sanitize(val)
}

func (srv *AppServer) SanitizeHtmlInput(val string) string {
	return srv.sanitizer.SanitizeHTML.Sanitize(val)
}
