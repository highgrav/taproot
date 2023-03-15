package workers

import "highgrav/taproot/v1/common"

type WorkRequest struct {
	Type string
	ID   string
	Data any
}

func NewWorkRequest(msgType string, t any) *WorkRequest {
	return &WorkRequest{
		Type: msgType,
		ID:   common.CreateRandString(16),
		Data: t,
	}
}
