package taproot

import (
	"errors"
	"github.com/highgrav/taproot/v1/workers"
)

func (srv *AppServer) AddWorkHandler(workType string, fn workers.WorkHandler) error {
	if srv.WorkHub == nil {
		return errors.New("work hub is not initialized")
	}
	srv.WorkHub.AddWorkFunc(workType, fn)
	return nil
}

func (srv *AppServer) AddWorkResultsHandler(workType string, fn workers.ResultHandler) error {
	if srv.WorkHub == nil {
		return errors.New("work hub is not initialized")
	}
	srv.WorkHub.AddResultsFunc(workType, fn)
	return nil
}

func (srv *AppServer) AddWorkHandlers(workType string, workFn workers.WorkHandler, resFn workers.ResultHandler) error {
	err := srv.AddWorkHandler(workType, workFn)
	if err != nil {
		return err
	}
	err = srv.AddWorkResultsHandler(workType, resFn)
	if err != nil {
		return err
	}
	return nil
}

func (srv *AppServer) StartWork(workType string, data any) (string, error) {
	if srv.WorkHub == nil {
		return "", errors.New("work hub is not initialized")
	}
	wr := workers.NewWorkRequest(workType, data)
	srv.WorkHub.Enqueue(wr)

	return wr.ID, nil
	return "", nil
}
