package session

import (
	"bytes"
	"encoding/gob"
)

type DefaultCodec struct {
}

func (codec DefaultCodec) Encode(obj any) ([]byte, error) {
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	err := enc.Encode(obj)
	if err != nil {
		return []byte{}, err
	}
	return m.Bytes(), nil
}

/*
Decode() decodes a byte array to an object. Note that decodedObj must be a pointer to a variable
*/
func (codec DefaultCodec) Decode(encodedObj []byte, decodedObj any) (any, error) {
	p := bytes.NewBuffer(encodedObj)
	dec := gob.NewDecoder(p)
	err := dec.Decode(decodedObj)
	if err != nil {
		return decodedObj, err
	}
	return decodedObj, nil
}
