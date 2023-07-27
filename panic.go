package taproot

import (
	"errors"
	"github.com/highgrav/taproot/common"
	"os"
	"path/filepath"
)

func (srv *AppServer) DumpStackTrace(trace string) (string, error) {
	if srv.Config.PanicStackTraceDirectory == "" {
		return "", errors.New("stack trace directory not configured")
	}
	randName := common.CreateRandString(16)
	randName = randName + ".trace"
	fileName := filepath.Join(srv.Config.PanicStackTraceDirectory, randName)
	err := os.WriteFile(fileName, []byte(trace), 0644)
	if err != nil {
		return "", err
	}
	return fileName, nil
}
