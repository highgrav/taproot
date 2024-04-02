package taproot

import (
	"context"
	"errors"
	"github.com/highgrav/taproot/logging"
	"github.com/highgrav/taproot/workers"
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
	err := srv.WorkHub.Enqueue(wr)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "TAPROOT", "error", "Failed to enqueue work '" + workType + "': " + err.Error())
	}
	return wr.ID, nil
	return "", nil
}
