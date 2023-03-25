package pagecache

import "time"

type PageCacheEvacFn func(id string)

type PageCacheEntry struct {
	ID    string
	Data  string
	timer *time.Timer
}

func NewPageCacheEntry(id, data string, evacDuration int, evacFn PageCacheEvacFn) *PageCacheEntry {
	pce := &PageCacheEntry{
		ID:    id,
		Data:  data,
		timer: time.NewTimer(time.Duration(evacDuration) * time.Second),
	}
	go func() {
		<-pce.timer.C
		evacFn(pce.ID)
	}()
	return pce
}
