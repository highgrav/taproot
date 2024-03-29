package pagecache

import (
	"context"
	"github.com/highgrav/taproot/logging"
)

type PageCache struct {
	Metrics *PageCacheMetrics
	cache   map[string]*PageCacheEntry
}

func NewPageCache() *PageCache {
	pc := &PageCache{
		Metrics: &PageCacheMetrics{},
		cache:   make(map[string]*PageCacheEntry),
	}
	return pc
}

func (pc *PageCache) Get(id string) (string, bool) {
	pce, ok := pc.cache[id]
	if !ok {
		return "", ok
	}
	return pce.Data, ok
}

func (pc *PageCache) expire(id string) {
	// TODO -- flags as a race condition
	delete(pc.cache, id)
}

func (pc *PageCache) Put(id, data string, secsToKeep int) {
	pce := NewPageCacheEntry(id, data, secsToKeep, pc.expire)
	pc.cache[id] = pce
	logging.LogToDeck(context.Background(), "info", "CACHE", "info", "adding "+id+" to page cache")
}

func (pc *PageCache) Flush(id string) {
	delete(pc.cache, id)
	logging.LogToDeck(context.Background(), "info", "CACHE", "info", "removing "+id+" from page cache")
}
