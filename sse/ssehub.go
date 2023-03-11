package sse

import (
	"strconv"
	"strings"
)

const SSE_MIMETYPE string = "text/event-stream"
const SSE_LAST_EVENT_SEEN_HEADER string = "Last-Event-ID"

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

type SSEHub struct {
	Name string
	// We assume that a constant key (ideally user ID, persistent session ID, etc.) is used here
	conns map[string][]chan SSEEvent
	acts  chan func() // prevents logical conflicts by single-threading operations
}

func (hub *SSEHub) runInternalActions() {
	for act := range hub.acts {
		act()
	} // infinite loop
}

func (hub *SSEHub) AddClient(clientId string, clientChan chan SSEEvent) {
	hub.acts <- func() {
		if _, ok := hub.conns[clientId]; !ok {
			hub.conns[clientId] = make([]chan SSEEvent, 0)
		}
		hub.conns[clientId] = append(hub.conns[clientId], clientChan)
	}
}

func (hub *SSEHub) RemoveClient(clientId string, clientChan chan SSEEvent) {
	if _, ok := hub.conns[clientId]; !ok {
		close(clientChan)
		return
	}
	go func() {
		for range clientChan {
		}
	}()
	hub.acts <- func() {
		tmpChs := make([]chan SSEEvent, 0)
		for _, c := range hub.conns[clientId] {
			if c != clientChan {
				tmpChs = append(tmpChs, c)
			}
		}
		// clean up if necessary
		if len(tmpChs) == 0 {
			delete(hub.conns, clientId)
		} else {
			hub.conns[clientId] = tmpChs
		}
		close(clientChan)
	}
}

func (hub *SSEHub) WriteOne(clientId string, msg SSEEvent) {
	hub.acts <- func() {
		chs, ok := hub.conns[clientId]
		if ok {
			for _, ch := range chs {
				ch <- msg
			}
		}
	}
}

func (broker *SSEHub) WriteMany(clientIds []string, msg SSEEvent) {
	broker.acts <- func() {
		for _, id := range clientIds {
			chs, ok := broker.conns[id]
			if ok {
				for _, ch := range chs {
					ch <- msg
				}
			}
		}
	}
}

func (broker *SSEHub) WriteAll(msg SSEEvent) {
	broker.acts <- func() {
		for _, v := range broker.conns {
			for _, c := range v {
				c <- msg
			}
		}
	}
}

func New(name string) *SSEHub {
	broker := &SSEHub{
		Name:  name,
		conns: make(map[string][]chan SSEEvent),
		acts:  make(chan func()),
	}
	go broker.runInternalActions()
	return broker
}
