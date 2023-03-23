package session

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
Decode() decodes a byte array to an object. decodedObj *must* be a pointer to an object
*/
func (codec DefaultCodec) Decode(encodedObj []byte, decodedObj any) (any, error) {
	p := bytes.NewBuffer(encodedObj)
	err := gob.NewDecoder(p).Decode(decodedObj)
	if err != nil {
		fmt.Println("ERROR DECODING")
		return decodedObj, err
	}
	return decodedObj, nil
}
