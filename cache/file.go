package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type (
	// FileCacheItem contains the cached data and expire time.
	FileCacheItem struct {
		Content interface{}
		Expire  time.Time
	}

	// FileCache is the file cache adapter.
	FileCache struct {
		Path string
		Ext  string
	}
)

// NewFileCache instantiate a new FileCache.
func NewFileCache() Cache {
	return &FileCache{}
}

// Hash the file name with MD5
func (fc *FileCache) getMD5Hash(name string) string {
	m := md5.New()
	m.Write([]byte(name))
	return filepath.Join(fc.Path, fmt.Sprintf("%s", hex.EncodeToString(m.Sum(nil))))
}

func fileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func gobEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(data []byte, item *FileCacheItem) error {
	buf := bytes.NewBuffer(data)
	return gob.NewDecoder(buf).Decode(&item)
}

// Init initialize the cache adapter with provided config string.
func (fc *FileCache) Init(config string) error {

	fc.Path = "cache"
	fc.Ext = ".bin"

	// Check if file cache path exists
	if ok := fileExists(fc.Path); !ok {
		_ = os.MkdirAll(fc.Path, os.ModePerm)
	}

	return nil
}

// Get cached value by key.
func (fc *FileCache) Get(key string) interface{} {
	filename := fc.getMD5Hash(key) + fc.Ext

	// Read file content
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}

	var item FileCacheItem
	gobDecode(data, &item)

	// If expired, removes the file and return an empty string
	if item.Expire.Before(time.Now()) {
		os.Remove(filename)
		return ""
	}

	return item.Content
}

// GetMulti values of Get.
func (fc *FileCache) GetMulti(keys []string) []interface{} {
	var out []interface{}
	for _, key := range keys {
		out = append(out, fc.Get(key))
	}
	return out
}

// Put sets the cache value with key and timeout.
func (fc *FileCache) Put(key string, value interface{}, timeout time.Duration) error {
	gob.Register(value)

	item := FileCacheItem{Content: value}
	item.Expire = time.Now().Add(timeout)

	data, err := gobEncode(item)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fc.getMD5Hash(key)+fc.Ext, data, 0644)
}

// Delete cached value by key.
func (fc *FileCache) Delete(key string) error {
	filename := fc.getMD5Hash(key) + fc.Ext
	if ok := fileExists(filename); ok {
		return os.Remove(filename)
	}
	return nil
}

// Exists check if cached value exists.
func (fc *FileCache) Exists(key string) bool {
	return fileExists(fc.getMD5Hash(key) + fc.Ext)
}

// ClearAll removes all cached values.
func (fc *FileCache) ClearAll() error {
	files, err := filepath.Glob(fc.Path + "/*" + fc.Ext)
	if err != nil {
		return err
	}

	for _, file := range files {
		os.Remove(file)
	}

	os.Remove(fc.Path)
	return nil
}
