package common

import (
	"sync"
	"time"
)

type KVCache[T any] struct {
	sync.Mutex
	cache map[string]*kvCacheEntry[T]
}

type kvCacheEntry[T any] struct {
	value     T
	createdOn time.Time
	lastSeen  time.Time
}

func (kv *KVCache[T]) Set(key string, val T) {
	kv.Lock()
	defer kv.Unlock()
	kv.cache[key] = &kvCacheEntry[T]{
		createdOn: time.Now(),
		lastSeen:  time.Now(),
	}
}

func (kv *KVCache[T]) Get(key string) (T, bool) {
	if v, ok := kv.cache[key]; ok {
		v.lastSeen = time.Now()
		return v.value, true
	}
	var t T
	return t, false
}

// Good defaults here are 15_000, 300_000, 3_600_000
func New[T any](sweepAt, staleAt, staleAbsoluteAt time.Duration) *KVCache[T] {
	kv := &KVCache[T]{
		cache: make(map[string]*kvCacheEntry[T]),
	}

	go func() {
		for true {
			time.Sleep(sweepAt)
			var marked []string = make([]string, len(kv.cache)/5) // arbitrary size
			for k, v := range kv.cache {
				if time.Now().After(v.createdOn.Add(staleAbsoluteAt)) || v.lastSeen.Before(time.Now().Add(-1*staleAt)) {
					marked = append(marked, k)
				}
			}

			for _, m := range marked {
				kv.Lock()
				delete(kv.cache, m)
				kv.Unlock()
			}
		}
	}()

	return kv
}
