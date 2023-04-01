package sse

import (
	"strconv"
	"strings"
)

type SSEFieldType string

const (
	SSEFIELD_COMMENT  SSEFieldType = ": "
	SSEFIELD_ID       SSEFieldType = "id: "
	SSEFIELD_EVENT    SSEFieldType = "event: "
	SSEFIELD_DATA     SSEFieldType = "data: "
	SSEFIELD_RETRY    SSEFieldType = "retry: "
	SSEFIELD_DISPATCH SSEFieldType = "\n"
)

type SSEEvent struct {
	UserID    string
	ID        string
	EventType string
	Data      []string
	Retry     int
}

func (evt *SSEEvent) Dispatch() string {
	sb := strings.Builder{}

	if evt.ID != "" {
		sb.Write([]byte(string(SSEFIELD_ID) + evt.ID + "\n"))
	}
	if evt.EventType != "" {
		sb.Write([]byte(string(SSEFIELD_EVENT) + evt.EventType + "\n"))
	}
	if len(evt.Data) > 0 {
		for _, v := range evt.Data {
			sb.Write([]byte(string(SSEFIELD_DATA) + v + "\n"))
		}
	}
	if evt.Retry > 0 {
		sb.Write([]byte(string(SSEFIELD_RETRY) + strconv.Itoa(evt.Retry) + "\n"))
	}

	sb.Write([]byte(SSEFIELD_DISPATCH))
	return sb.String()
}
