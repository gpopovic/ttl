// Package ttl implements routines to manage a TTL cache.
package ttl

import (
	"sync"
	"time"
)

// item is a cache item with a timer.
type item struct {
	Start    time.Time
	Timer    *time.Timer
	Duration time.Duration
	Value    interface{}
}

// Cache is a TTL cache, safe for concurrent access.
type Cache struct {
	Duration   time.Duration
	ResetOnAdd bool
	mu         sync.Mutex
	items      map[string]*item
}

// New returns a new cache using duration as the cache time limit.
func New(duration time.Duration) *Cache {
	return &Cache{
		Duration: duration,
		items:    make(map[string]*item),
	}
}

// Len gets the length of items in the cache.
func (cache *Cache) Len() int {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	return len(cache.items)
}

// Add adds an item to the cache, starting a timer to remove if it doesn't exist. If it does exist
// and ResetOnAdd is true the keys timer will reset to the original duration.
func (cache *Cache) Add(key string, value interface{}) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	i, ok := cache.items[key]
	if !ok {
		i = &item{time.Now(), time.NewTimer(cache.Duration), cache.Duration, value}

		go func() {
			select {
			case <-i.Timer.C:
				cache.mu.Lock()
				defer cache.mu.Unlock()

				delete(cache.items, key)
			}
		}()
	} else {
		i.Value = value
	}

	if ok && cache.ResetOnAdd {
		ok = i.Timer.Reset(i.Duration)
		if !ok {
			i.Timer = time.NewTimer(i.Duration)
		}

		i.Start = time.Now()
	}

	cache.items[key] = i
}

// Get retrieves an item from the cache, the second boolean value denotes if the key exists.
func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	item, ok := cache.items[key]
	if !ok {
		return nil, ok
	}

	return item.Value, ok
}

// TTL returns a duration for the amount of time left the key has. The second boolean denotes
// if the key exists.
func (cache *Cache) TTL(key string) (time.Duration, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	item, ok := cache.items[key]
	if !ok {
		return 0, ok
	}

	return item.Start.Add(item.Duration).Sub(time.Now()), ok
}
