package logging

import (
	"context"
	"fmt"
	"github.com/google/deck"
	"github.com/highgrav/taproot/constants"
	"strings"
	"time"
)

func LogToDeck(ctx context.Context, crit, app, level, msg string) {

	corrId, ok := ctx.Value(constants.HTTP_CONTEXT_CORRELATION_KEY).(string)
	if !ok {
		corrId = "-"
	}
	sessionId, ok := ctx.Value(constants.HTTP_CONTEXT_SESSION_KEY).(string)
	if !ok {
		sessionId = "-"
	}

	val := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", app, level, corrId, sessionId, msg)
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

func LogString(crit, msg string) {
	if strings.ToLower(crit) == "info" {
		deck.Info(msg)
	} else if strings.ToLower(crit) == "error" {
		deck.Error(msg)
	} else if strings.ToLower(crit) == "fatal" {
		deck.Fatal(msg)
	} else if strings.ToLower(crit) == "warn" || strings.ToLower(crit) == "warning" {
		deck.Warning(msg)
	} else {
		deck.Info(msg)
	}
}

func LogW3CRequest(crit string, reqTime time.Time, clientIp string, ctx context.Context, rMethod, rURL, userId string) {
	customTimeFormat := "2006-01-02T15:04:05.000-07:00"
	corrId, ok := ctx.Value(constants.HTTP_CONTEXT_CORRELATION_KEY).(string)
	if !ok {
		corrId = ""
	}

	val := fmt.Sprintf("REQ\t%s\t%s\t%s\t%s\t%s\t%s\t\t\t\n", clientIp, corrId, userId, reqTime.Format(customTimeFormat), rMethod, rURL)
	LogString(crit, val)
}

func LogW3CResponse(crit string, reqTime time.Time, clientIp string, ctx context.Context, rMethod, rURL string, rCode int, rBytesWritten int, userId string) {
	customTimeFormat := "2006-01-02T15:04:05.000-07:00"
	corrId, ok := ctx.Value(constants.HTTP_CONTEXT_CORRELATION_KEY).(string)
	if !ok {
		corrId = ""
	}

	val := fmt.Sprintf("RES\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\n", clientIp, corrId, userId, time.Now().Format(customTimeFormat), time.Now().Sub(reqTime).String(), rMethod, rURL, rCode, rBytesWritten)
	LogString(crit, val)
}

func LogTaskStart(crit string, taskType string, taskId string, msg string) {
	customTimeFormat := "2006-01-02T15:04:05.000-07:00"
	tm := time.Now().Format(customTimeFormat)
	val := fmt.Sprintf("TSKBGN\t%s\t%s\t%s\t%s\n", taskType, taskId, tm, msg)
	LogString(crit, val)
}

func LogTaskEnd(crit string, taskType string, taskId string, result string, msg string) {
	customTimeFormat := "2006-01-02T15:04:05.000-07:00"
	tm := time.Now().Format(customTimeFormat)
	val := fmt.Sprintf("TSKEND\t%s\t%s\t%s\t%s\t%s\n", taskType, taskId, tm, result, msg)
	LogString(crit, val)
}
