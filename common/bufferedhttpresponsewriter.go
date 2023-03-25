package common

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"sync"
)

type BufferedHttpResponseWriter struct {
	sync.Mutex
	http.ResponseWriter
	Buf           bytes.Buffer
	Code          int
	IsWritingHttp bool
	IsClosed      bool
	Result        *strings.Builder
}

func NewBufferedHttpResponseWriter(w http.ResponseWriter) *BufferedHttpResponseWriter {
	bw := &BufferedHttpResponseWriter{
		Mutex:          sync.Mutex{},
		ResponseWriter: w,
		Buf:            bytes.Buffer{},
		Code:           200,
		IsWritingHttp:  false,
		IsClosed:       false,
		Result:         new(strings.Builder),
	}
	return bw
}

func (bw *BufferedHttpResponseWriter) Flush() (int, error) {
	bw.Lock()
	defer bw.Unlock()
	if bw.IsClosed {
		return -1, errors.New("httpwriter is closed")
	}
	if !bw.IsWritingHttp {
		bw.ResponseWriter.WriteHeader(bw.Code)
		bw.IsWritingHttp = true
	}
	i, err := bw.ResponseWriter.Write(bw.Buf.Bytes())
	bw.Result.Write(bw.Buf.Bytes())
	if err == nil {
		bw.Buf.Truncate(0)
	}
	return i, err

}

func (bw *BufferedHttpResponseWriter) Close() {
	bw.Lock()
	defer bw.Unlock()
	bw.IsClosed = true
}

func (bw *BufferedHttpResponseWriter) String() string {
	return bw.Result.String()
}

func (bw *BufferedHttpResponseWriter) Write(b []byte) (int, error) {
	return bw.Buf.Write(b)
}

func (bw *BufferedHttpResponseWriter) WriteHeaderPair(name, val string) {
	bw.ResponseWriter.Header().Set(name, val)
}

func (bw *BufferedHttpResponseWriter) WriteHeader(code int) {
	bw.Code = code
}
