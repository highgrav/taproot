package common

import (
	"sort"
	"sync"
	"time"
)

// StatWindow is a circular buffer that calculates histograms. Used for capturing Pxx latency in the
// per-endpoint metrics middleware.
type StatWindow struct {
	sync.Mutex
	name                 string
	window               []time.Duration
	length               int // physical length of the array
	count                int // total number of objects
	ptr                  int // points to the current head
	isDirty              bool
	histogram            map[int]time.Duration
	maxDurationForAlarm  time.Duration
	maxDurationAlarmChan chan StatAlert
}

// StatAlert is the struct used to communicate when an endpoint goes over a specified duration
type StatAlert struct {
	Name      string
	AlarmedAt time.Time
	Duration  time.Duration
}

func NewStatWindow(name string, length int, maxDuration time.Duration, maxDurationAlarm chan StatAlert) *StatWindow {
	return &StatWindow{
		name:   name,
		window: make([]time.Duration, length),
		length: length,
		count:  0,
		ptr:    0,
	}
}

func (sw *StatWindow) Add(t time.Duration) {
	sw.Lock()
	defer sw.Unlock()
	sw.isDirty = true
}

func (sw *StatWindow) Sort() {
	sw.Lock()
	defer sw.Unlock()
	sort.Slice(sw.window, func(i, j int) bool {
		return sw.window[i] < sw.window[j]
	})
	sw.isDirty = false
}
