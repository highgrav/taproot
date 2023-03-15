package mq

import (
	"github.com/joncrlsn/dque"
)

/*
The MQueue is a centralized message queue that takes messages and dispatches them to specified functions.
Messages are serialized to disk and are durable between restarts.
*/
type MQueue struct {
	Errors   chan error
	queue    *dque.DQue
	handlers map[string][]QueueHandler
}

// Places a message on the queue for processing
func (mq *MQueue) Enqueue(msg *QueueMsg) error {
	return mq.queue.Enqueue(msg)
}

// Adds a function to process a specified message type
func (mq *MQueue) AddFunc(msgType string, fn QueueHandler) {
	if _, ok := mq.handlers[msgType]; !ok {
		mq.handlers[msgType] = make([]QueueHandler, 0)
	}
	mq.handlers[msgType] = append(mq.handlers[msgType], fn)
}

// goroutine to process incoming messages and dispatch them
func (mq *MQueue) processMsgs() error {
	for {
		msg, err := mq.queue.DequeueBlock()
		if err != nil {
			mq.Errors <- err
		} else {
			fns, ok := mq.handlers[msg.(QueueMsg).Type]
			if ok {
				for _, fn := range fns {
					err = fn(msg.(QueueMsg))
					if err != nil {
						mq.Errors <- err
					}
				}
			}
		}
	}
}

func msgBuilder() interface{} {
	return &QueueMsg{}
}

// Creates a new MQ
func New(name string, saveDir string, segmentSz int) (*MQueue, error) {
	dq, err := dque.NewOrOpen(name, saveDir, segmentSz, msgBuilder)
	if err != nil {
		return nil, err
	}
	mq := &MQueue{
		Errors:   make(chan error),
		queue:    dq,
		handlers: make(map[string][]QueueHandler),
	}

	// start infinite loop to process messages
	go mq.processMsgs()

	return mq, nil
}
