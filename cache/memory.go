package cache

import (
	"sync"
	"time"
)

type (
	// MemoryCacheItem contains the cached data and expire time.
	MemoryCacheItem struct {
		Content interface{}
		Expire  time.Time
	}

	// MemoryCache is memory cache adapter.
	MemoryCache struct {
		sync.RWMutex
		items map[string]*MemoryCacheItem

		dur   time.Duration
		Every int // run an expiration check Every clock time
	}
)

// NewMemoryCache instantiate a new MemoryCache.
func NewMemoryCache() Cache {
	return &MemoryCache{items: make(map[string]*MemoryCacheItem)}
}

// Init initialize the cache adapter with provided config string.
func (mc *MemoryCache) Init(config string) error {
	return nil
}

// Get cached value by key.
func (mc *MemoryCache) Get(key string) interface{} {
	mc.RLock()
	item := mc.items[key]
	if item == nil {
		mc.RUnlock()
		return ""
	}

	// If expired, removes the value from memory and return an empty string
	if item.Expire.Before(time.Now()) {
		mc.Delete(key)
		mc.RUnlock()
		return ""
	}

	mc.RUnlock()
	return item.Content
}

// GetMulti values of Get.
func (mc *MemoryCache) GetMulti(keys []string) []interface{} {
	var out []interface{}
	for _, key := range keys {
		out = append(out, mc.Get(key))
	}
	return out
}

// Put sets the cache value with key and timeout.
func (mc *MemoryCache) Put(key string, value interface{}, timeout time.Duration) error {
	mc.Lock()
	mc.items[key] = &MemoryCacheItem{
		Content: value,
		Expire:  time.Now().Add(timeout),
	}
	mc.Unlock()
	return nil
}

// Delete cached value by key.
func (mc *MemoryCache) Delete(key string) error {
	mc.Lock()
	delete(mc.items, key)
	mc.Unlock()
	return nil
}

// Exists check if cached value exists.
func (mc *MemoryCache) Exists(key string) bool {
	mc.RLock()
	defer mc.RUnlock()
	if item, ok := mc.items[key]; ok {
		if item.Expire.Before(time.Now()) {
			mc.Delete(key)
			return false
		}
		return true
	}
	return false
}

// ClearAll removes all cached values.
func (mc *MemoryCache) ClearAll() error {
	mc.Lock()
	mc.items = make(map[string]*MemoryCacheItem)
	mc.Unlock()
	return nil
}

// func init() {
// 	Register("memory", NewMemoryCache)
// }
