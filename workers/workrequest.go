package workers

import "github.com/highgrav/taproot/common"

type WorkRequest struct {
	Type string
	ID   string
	Data any
}

func NewWorkRequest(msgType string, t any) *WorkRequest {
	return &WorkRequest{
		Type: msgType,
		ID:   common.CreateRandString(24),
		Data: t,
	}
}
