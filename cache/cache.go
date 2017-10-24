package cache

import (
	"fmt"
	"time"
)

type (
	// Cache contains the base cache adapter.
	Cache interface {
		// Init initialize the cache adapter with provided config string.
		Init(string) error
		// Get cached value by key.
		Get(string) interface{}
		// GetMulti values of Get.
		GetMulti([]string) []interface{}
		// Put sets the cache value with key and timeout.
		Put(string, interface{}, time.Duration) error
		// Delete cached value by key.
		Delete(string) error
		// Exists check if cached value exists.
		Exists(string) bool
		// ClearAll removes all cached values.
		ClearAll() error
	}
	// Adapter is the cache instance.
	Adapter func() Cache
)

var adapters = make(map[string]Adapter)

// Register adds the cache adapter by name
func Register(name string, adapter Adapter) error {
	if name == "" {
		return fmt.Errorf("cache register: name is empty")
	}

	if adapter == nil {
		return fmt.Errorf("cache register: adapter is nil")
	}

	if _, ok := adapters[name]; ok {
		return fmt.Errorf("cache register: adapter %s already registered", name)
	}

	adapters[name] = adapter

	return nil
}

// NewCache creates the cache instance using the provided adapter and config string (must contain a valid JSON string).
func NewCache(name, config string) (Cache, error) {
	instance, ok := adapters[name]
	if !ok {
		return nil, fmt.Errorf("cache: unknown adapter %s", name)
	}

	adapter := instance()
	err := adapter.Init(config)
	if err != nil {
		return nil, err
	}

	return adapter, nil
}
