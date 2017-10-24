// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AnUnnamedProject/framework/utils"
)

type (
	// FileSessionProvider contains the cookie store.
	FileSessionProvider struct {
		sync.RWMutex
		values map[string]interface{}

		sessionID string
		config    *fileConfig
	}

	fileConfig struct {
		SessionConfig
		SavePath string `json:"save_path"`
	}
)

// NewFileSessionProvider returns a new file session provider.
func NewFileSessionProvider(config string) (Session, error) {
	var err error
	fsp := &FileSessionProvider{}

	fsp.config = &fileConfig{}
	err = json.Unmarshal([]byte(config), fsp.config)
	if err != nil {
		return nil, err
	}

	if fsp.config.SavePath == "" {
		return nil, fmt.Errorf("file session: empty SavePath")
	}

	fi, err := os.Stat(fsp.config.SavePath)
	if err != nil {
		return nil, fmt.Errorf("file session: SavePath error %v\n", err)
	}

	// Check if the directory is valid
	if !fi.IsDir() {
		return nil, fmt.Errorf("file session: %s is not a directory", fsp.config.SavePath)
	}

	// If SavePath is not absolute, get current absolute path
	if !filepath.IsAbs(fsp.config.SavePath) {
		var wd string
		wd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("file session: Getwd error %v\n", err)
		}

		fsp.config.SavePath = filepath.Join(wd, fsp.config.SavePath)
	}

	// Check if SavePath directory exists
	err = os.MkdirAll(fsp.config.SavePath, 0660)
	if err != nil {
		return nil, err
	}

	// Start GC
	fsp.GC()

	return fsp, nil
}

// Init checks if session id exists and return it.
func (fsp *FileSessionProvider) Init(rw http.ResponseWriter, r *http.Request) (Session, error) {
	fsp.RLock()
	defer fsp.RUnlock()

	// Retreive session id
	sessionID := GetSessionID(r, fsp.config.Name)

	// Session found, decode and set values
	if sessionID != "" {
		var fh *os.File
		var err error
		sessionID, err = DecodeCookie([]byte(fsp.config.Key), fsp.config.MaxLifetime, sessionID, r)

		if err != nil {
			fmt.Printf("file session init returned an error while decoding cookie: %v\n", err)

			fsp.sessionID = sessionID

			// Destroy old session
			_ = fsp.Destroy()

			// Create new session
			goto newSession
		}

		if sessionID == "" {
			goto newSession
		}

		// Read session from file
		_, err = os.Stat(filepath.Join(fsp.config.SavePath, sessionID))
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("file session init returned an error while checking for file: %v\n", err)
			return nil, err
		}
		fh, err = os.OpenFile(filepath.Join(fsp.config.SavePath, sessionID), os.O_RDWR|os.O_CREATE, 0660)
		if err != nil {
			return nil, err
		}

		// Update file access / modification time for GC
		_ = os.Chtimes(filepath.Join(fsp.config.SavePath, sessionID), time.Now(), time.Now())

		var values map[string]interface{}
		b, err := ioutil.ReadAll(fh)
		if err != nil {
			return nil, err
		}

		if len(b) == 0 {
			values = make(map[string]interface{})
		}

		err = utils.GobDecode(b, &values)
		if err != nil {
			return nil, err
		}
		_ = fh.Close()

		return &FileSessionProvider{sessionID: sessionID, config: fsp.config, values: values}, nil
	}

newSession:
	fsp.values = make(map[string]interface{})

	// Create new session
	fsp.sessionID = NewSessionID()

	encoded := EncodeCookie([]byte(fsp.config.Key), fsp.sessionID, r)

	cookie := &http.Cookie{
		Name:     fsp.config.Name,
		Value:    url.QueryEscape(encoded),
		Path:     "/",
		Domain:   fsp.config.Domain,
		HttpOnly: fsp.config.HTTPOnly,
		Secure:   false,
		MaxAge:   fsp.config.MaxLifetime,
	}

	if fsp.config.MaxLifetime > 0 {
		cookie.Expires = time.Now().Add(time.Duration(fsp.config.MaxLifetime) * time.Second)
	}

	http.SetCookie(rw, cookie)
	r.AddCookie(cookie)

	return fsp, nil
}

// Release writes the cookie back to the http response cookie
func (fsp *FileSessionProvider) Release(rw http.ResponseWriter, req *http.Request) {
	fsp.Lock()
	defer fsp.Unlock()

	// session ID does not exist, destroyed by user
	if fsp.sessionID == "" {
		return
	}

	b, err := utils.GobEncode(fsp.values)
	if err != nil {
		return
	}

	var fh *os.File
	_, err = os.Stat(filepath.Join(fsp.config.SavePath, fsp.sessionID))
	if err != nil && !os.IsNotExist(err) {
		return
	}
	fh, err = os.OpenFile(filepath.Join(fsp.config.SavePath, fsp.sessionID), os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return
	}

	_ = fh.Truncate(0)
	_, _ = fh.Seek(0, 0)
	_, _ = fh.Write(b)
	_ = fh.Close()
}

// Set value to cookie session.
func (fsp *FileSessionProvider) Set(key string, value interface{}) error {
	fsp.Lock()
	fsp.values[key] = value
	fsp.Unlock()
	return nil
}

// Get value from cookie session.
func (fsp *FileSessionProvider) Get(key string) interface{} {
	fsp.RLock()
	if value, ok := fsp.values[key]; ok {
		fsp.RUnlock()
		return value
	}

	fsp.RUnlock()
	return nil
}

// Delete value in cookie session.
func (fsp *FileSessionProvider) Delete(key string) error {
	fsp.Lock()
	delete(fsp.values, key)
	fsp.Unlock()
	return nil
}

// Destroy entire session
func (fsp *FileSessionProvider) Destroy() error {
	_, err := os.Stat(filepath.Join(fsp.config.SavePath, fsp.sessionID))
	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(filepath.Join(fsp.config.SavePath, fsp.sessionID))
	if err != nil {
		return err
	}

	fsp.sessionID = ""

	return nil
}

// SessionID Return id of this cookie session
func (fsp *FileSessionProvider) SessionID() string {
	return fsp.sessionID
}

// GC clean expired sessions.
func (fsp *FileSessionProvider) GC() {
	time.AfterFunc(time.Duration(fsp.config.MaxLifetime)*time.Second, func() {
		dsa := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if (info.ModTime().Unix() + int64(fsp.config.MaxLifetime)) < time.Now().Unix() {
				_ = os.Remove(path)
			}

			return nil
		}
		_ = filepath.Walk(fsp.config.SavePath, dsa)
		fsp.GC()
	})
}

// Handler
func (fsp *FileSessionProvider) Handler(c *Context) {
	var err error
	c.Session, err = fsp.Init(c.Response, c.Request.Request)
	if err != nil {
		// TODO: Throw error 500
		c.Error(500, fmt.Errorf("Internal error"))
		Log.Error(err)
		return
	}
	// Write back session id to client
	c.Response.Before(func(bw ResponseWriter) {
		c.Session.Release(bw, c.Request.Request)
	})
}
