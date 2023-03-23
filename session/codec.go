package session

type ICodec interface {
	Encode(any) ([]byte, error)
	Decode(encodedObj []byte, decodedObj any) (any, error)
}
