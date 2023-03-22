package common

import (
	"bytes"
	"net/http"
)

type BufferedResponseWriter struct {
	http.ResponseWriter
	Buf  bytes.Buffer
	Code int
}

func (bw *BufferedResponseWriter) Write(b []byte) (int, error) {
	return bw.Buf.Write(b)
}

func (bw *BufferedResponseWriter) WriteHeader(code int) {
	bw.Code = code
}
