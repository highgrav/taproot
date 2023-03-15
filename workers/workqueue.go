package workers

import (
	"github.com/joncrlsn/dque"
)

/*
The WorkQueue is a centralized message queue that takes messages and dispatches them to specified functions.
Messages are serialized to disk and are durable between restarts.
*/
type WorkQueue struct {
	Status         chan WorkStatusReport
	queue          *dque.DQue
	workHandlers   map[string][]WorkHandler
	resultHandlers map[string][]ResultHandler
}

// Places a message on the queue for processing
func (wq *WorkQueue) Enqueue(msg *WorkRequest) error {
	return wq.queue.Enqueue(msg)
}

// Adds a function to process a specified message type
func (wq *WorkQueue) AddWorkFunc(msgType string, fn WorkHandler) {
	if _, ok := wq.workHandlers[msgType]; !ok {
		wq.workHandlers[msgType] = make([]WorkHandler, 0)
	}
	wq.workHandlers[msgType] = append(wq.workHandlers[msgType], fn)
}

func (wq *WorkQueue) AddResultsFunc(msgType string, fn ResultHandler) {
	if _, ok := wq.resultHandlers[msgType]; !ok {
		wq.resultHandlers[msgType] = make([]ResultHandler, 0)
	}
	wq.resultHandlers[msgType] = append(wq.resultHandlers[msgType], fn)
}

func (wq *WorkQueue) processResults() {
	for {
		select {
		case res := <-wq.Status:
			hs := wq.resultHandlers[res.Type]
			for _, fn := range hs {
				go func() {
					fn(res)
				}()
			}
		}
	}
}

// goroutine to process incoming messages and dispatch them
func (wq *WorkQueue) processMsgs() {
	for {
		msg, err := wq.queue.DequeueBlock()
		if err != nil {
			wq.Status <- WorkStatusReport{
				Status: "failed to dequeue msg",
				Error:  err,
			}
		} else {
			fns, ok := wq.workHandlers[msg.(*WorkRequest).Type]
			if ok {
				for _, fn := range fns {
					go func() {
						res := fn(msg.(*WorkRequest))
						wq.Status <- res
					}()
				}
			}
		}
	}
}

func msgBuilder() interface{} {
	return &WorkRequest{}
}

// Creates a new MQ
func New(name string, saveDir string, segmentSz int) (*WorkQueue, error) {
	dq, err := dque.NewOrOpen(name, saveDir, segmentSz, msgBuilder)
	if err != nil {
		return nil, err
	}
	wq := &WorkQueue{
		Status:         make(chan WorkStatusReport),
		queue:          dq,
		workHandlers:   make(map[string][]WorkHandler),
		resultHandlers: make(map[string][]ResultHandler),
	}

	// start infinite loop to process messages and results
	go wq.processMsgs()
	go wq.processResults()
	return wq, nil
}
