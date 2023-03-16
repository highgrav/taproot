package cron

import (
	"errors"
	"github.com/gorhill/cronexpr"
	"highgrav/taproot/v1/logging"
	"sync"
	"time"
)

var ErrNamedJobAlreadyExists = errors.New("job already exists, please remove by name first")

/*
The CronHub schedules jobs using a simple cron syntax (see github.com/gorhill/cronexpr).
Unlike workers, cronjobs are not durable between restarts, so if a server is down, jobs may be missed.
*/
type CronHub struct {
	sync.Mutex
	Entries map[string]*CronEntry
	Pause   chan bool
	Done    chan bool
	Paused  bool
}

func New() *CronHub {
	ch := &CronHub{
		Mutex:   sync.Mutex{},
		Entries: make(map[string]*CronEntry, 0),
		Pause:   make(chan bool),
		Done:    make(chan bool),
	}

	go ch.loopForJobs()
	return ch
}

// Adds a job with a name
func (ch *CronHub) AddJob(name, schedule string, job CronJob) error {
	if _, ok := ch.Entries[name]; ok {
		return ErrNamedJobAlreadyExists
	}

	entry := &CronEntry{
		Name:        name,
		Malformed:   false,
		Schedule:    schedule,
		NextRunTime: time.Time{},
		Job:         job,
	}
	t, err := cronexpr.Parse(entry.Schedule)
	if err != nil {
		return err
	}
	entry.NextRunTime = t.Next(time.Now())
	ch.Lock()
	defer ch.Unlock()
	ch.Entries[name] = entry
	return nil
}

// Removes a job using its unique name
func (ch *CronHub) RemoveJob(name string) {
	ch.Lock()
	defer ch.Unlock()
	delete(ch.Entries, name)
}

func (ch *CronHub) scheduleAll() {
	ch.Lock()
	defer ch.Unlock()
	currTime := time.Now()
	for name, entry := range ch.Entries {
		t, err := cronexpr.Parse(entry.Schedule)
		if err == nil {
			entry.Malformed = true
			logging.LogToDeck("error", "CRON\t"+name+"\tMalformed cron entry for "+name+" ("+entry.Schedule+")")
			continue
		}
		entry.NextRunTime = t.Next(currTime)
	}
}

// Loops endlessly, looking for jobs to run every minute
func (ch *CronHub) loopForJobs() {
	logging.LogToDeck("info", "Cronjobs starting")
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for {
			select {
			case <-ch.Done:
				return
			case <-ch.Pause:
				ch.Paused = !ch.Paused
			case t := <-ticker.C:
				if !ch.Paused {
					ch.runJobs(t)
				}
			}
		}
	}()
}

func (ch *CronHub) runJobs(currTime time.Time) {
	ch.Lock()
	defer ch.Unlock()
	for _, entry := range ch.Entries {
		if entry.Malformed {
			continue
		}
		if currTime.After(entry.NextRunTime) {
			t, err := cronexpr.Parse(entry.Schedule)
			if err != nil {
				logging.LogToDeck("error", "CRON\t"+entry.Name+"\tMalformed cron entry for "+entry.Name+" ("+entry.Schedule+")")
				entry.Malformed = true
				continue
			}
			entry.NextRunTime = t.Next(time.Now())
			go entry.Job()
		}
	}
}
