package logging

import (
	"fmt"
	"github.com/google/deck"
	"strings"
	"time"
)

func LogToDeck(crit, val string) {
	if strings.ToLower(crit) == "info" {
		deck.Info(val)
	} else if strings.ToLower(crit) == "error" {
		deck.Error(val)
	} else if strings.ToLower(crit) == "fatal" {
		deck.Fatal(val)
	} else if strings.ToLower(crit) == "warn" || strings.ToLower(crit) == "warning" {
		deck.Warning(val)
	} else {
		deck.Info(val)
	}
}

func LogW3CRequest(crit string, reqTime time.Time, clientIp, corrId, rMethod, rURL string) {
	customTimeFormat := "2006-01-02T15:04:05.000-07:00"
	val := fmt.Sprintf("REQ\t%s\t%s\t-\t%s\t%s\t%s\t\t\t\n", clientIp, corrId, reqTime.Format(customTimeFormat), rMethod, rURL)
	LogToDeck(crit, val)
}

func LogW3CResponse(crit string, reqTime time.Time, clientIp, corrId, rMethod, rURL string, rCode int, rBytesWritten int) {
	customTimeFormat := "2006-01-02T15:04:05.000-07:00"
	val := fmt.Sprintf("RES\t%s\t%s\t-\t%s\t%s\t%s\t%s\t%d\t%d\n", clientIp, corrId, time.Now().Format(customTimeFormat), time.Now().Sub(reqTime).String(), rMethod, rURL, rCode, rBytesWritten)
	LogToDeck(crit, val)
}
