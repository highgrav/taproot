package taproot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DataEnvelope map[string]any

func (srv *Server) WriteJSON(w http.ResponseWriter, prettyPrint bool, status int, data DataEnvelope, headers http.Header) error {
	var js []byte
	var err error
	if prettyPrint {
		js, err = json.MarshalIndent(data, "", "\t")
	} else {
		js, err = json.Marshal(data)
	}
	if err != nil {
		return nil
	}

	js = append(js, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (srv *Server) ReadJSONFromBody(w http.ResponseWriter, r *http.Request, dst any) error {
	var maxBytes int64 = 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains invalid JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains invalid JSON (unexpected end of input)")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type (for field %q at character %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body is empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fn := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown field (for field %s)", fn)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not exceed %d bytes", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
		return nil
	}

	// catch any attempt to pass multiple JSON structs in`
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body contains multiple JSON values")
	}

	return nil
}
