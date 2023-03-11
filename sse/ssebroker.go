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

type SSEBroker struct {
	Name string
	// We assume that a constant key (ideally user ID, persistent session ID, etc.) is used here
	conns map[string][]chan SSEEvent
	acts  chan func() // prevents logical conflicts by single-threading operations
}

func (broker *SSEBroker) runInternalActions() {
	for act := range broker.acts {
		act()
	} // infinite loop
}

func (broker *SSEBroker) AddClient(clientId string, clientChan chan SSEEvent) {
	broker.acts <- func() {
		if _, ok := broker.conns[clientId]; !ok {
			broker.conns[clientId] = make([]chan SSEEvent, 0)
		}
		broker.conns[clientId] = append(broker.conns[clientId], clientChan)
	}
}

func (broker *SSEBroker) RemoveClient(clientId string, clientChan chan SSEEvent) {
	if _, ok := broker.conns[clientId]; !ok {
		close(clientChan)
		return
	}
	go func() {
		for range clientChan {
		}
	}()
	broker.acts <- func() {
		tmpChs := make([]chan SSEEvent, 0)
		for _, c := range broker.conns[clientId] {
			if c != clientChan {
				tmpChs = append(tmpChs, c)
			}
		}
		// clean up if necessary
		if len(tmpChs) == 0 {
			delete(broker.conns, clientId)
		} else {
			broker.conns[clientId] = tmpChs
		}
		close(clientChan)
	}
}

func (broker *SSEBroker) WriteOne(clientId string, msg SSEEvent) {
	broker.acts <- func() {
		chs, ok := broker.conns[clientId]
		if ok {
			for _, ch := range chs {
				ch <- msg
			}
		}
	}
}

func (broker *SSEBroker) WriteMany(clientIds []string, msg SSEEvent) {
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

func (broker *SSEBroker) WriteAll(msg SSEEvent) {
	broker.acts <- func() {
		for _, v := range broker.conns {
			for _, c := range v {
				c <- msg
			}
		}
	}
}

func New(name string) *SSEBroker {
	broker := &SSEBroker{
		Name:  name,
		conns: make(map[string][]chan SSEEvent),
		acts:  make(chan func()),
	}
	go broker.runInternalActions()
	return broker
}
