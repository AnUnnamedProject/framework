package cache

import (
	"testing"
	"time"
)

func TestFileCache(t *testing.T) {
	c, err := NewCache("file", "")
	if err != nil {
		t.Errorf("Unable to init FileCache: %v\n", err)
	}

	// Test PUT
	if err = c.Put("test", "test", 10*time.Second); err != nil {
		t.Errorf("Put: unable to write: %v\n", err)
	}

	if !c.Exists("test") {
		t.Error("Exists: unable to find cached value")
	}

	// Test GET
	if v := c.Get("test"); v != "test" {
		t.Error("Get: unable to get cached value")
	}

	// Test Delete
	err = c.Delete("test")
	if err != nil {
		t.Errorf("Delete: unable to delete cached value: %v\n", err)
	}
	if c.Exists("test") {
		t.Error("Delete: unable to delete cached value")
	}

	// Test GetMulti
	c.Put("test", "test", 10*time.Second)
	c.Put("test2", "test2", 10*time.Second)

	vv := c.GetMulti([]string{"test", "test2"})
	if len(vv) != 2 {
		t.Error("GetMulti: unable to retreive cached values")
	}
	if vv[0].(string) != "test" {
		t.Error("GetMulti: unable to retreive cached values")
	}
	if vv[1].(string) != "test2" {
		t.Error("GetMulti: unable to retreive cached values")
	}

	// Remove tests
	c.ClearAll()
}

func TestMemoryCache(t *testing.T) {
	c, err := NewCache("memory", "")
	if err != nil {
		t.Errorf("Unable to init MemoryCache: %v\n", err)
	}

	// Test PUT
	if err = c.Put("test", "test", 10*time.Second); err != nil {
		t.Errorf("Put: unable to write: %v\n", err)
	}

	if !c.Exists("test") {
		t.Error("Exists: unable to find cached value")
	}

	// Test GET
	if v := c.Get("test"); v != "test" {
		t.Error("Get: unable to get cached value")
	}

	// Test Delete
	err = c.Delete("test")
	if err != nil {
		t.Errorf("Delete: unable to delete cached value: %v\n", err)
	}
	if c.Exists("test") {
		t.Error("Delete: unable to delete cached value")
	}

	// Test GetMulti
	c.Put("test", "test", 10*time.Second)
	c.Put("test2", "test2", 10*time.Second)

	vv := c.GetMulti([]string{"test", "test2"})
	if len(vv) != 2 {
		t.Error("GetMulti: unable to retreive cached values")
	}
	if vv[0].(string) != "test" {
		t.Error("GetMulti: unable to retreive cached values")
	}
	if vv[1].(string) != "test2" {
		t.Error("GetMulti: unable to retreive cached values")
	}

	// Remove tests
	c.ClearAll()
}
