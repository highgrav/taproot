package common

import (
	"bytes"
	"net/http"
)

type BufferedHttpResponseWriter struct {
	http.ResponseWriter
	Buf  bytes.Buffer
	Code int
}

func (bw *BufferedHttpResponseWriter) Write(b []byte) (int, error) {
	return bw.Buf.Write(b)
}

func (bw *BufferedHttpResponseWriter) WriteHeader(code int) {
	bw.Code = code
}
