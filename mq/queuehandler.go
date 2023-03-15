package mq

type QueueHandler func(msg QueueMsg) error
