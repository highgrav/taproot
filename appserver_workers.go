package taproot

import (
	"errors"
	"highgrav/taproot/v1/workers"
)

func (srv *AppServer) StartWork(workType string, data any) (string, error) {
	if srv.WorkHub == nil {
		return "", errors.New("work hub is not initialized")
	}
	wr := workers.NewWorkRequest(workType, data)

	return wr.ID, nil
}
