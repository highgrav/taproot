package cron

import "time"

type CronJob func() error

type CronEntry struct {
	Name        string
	Malformed   bool
	Schedule    string
	NextRunTime time.Time
	Job         CronJob
}
