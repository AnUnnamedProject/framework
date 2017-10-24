// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"io/ioutil"

	"github.com/AnUnnamedProject/i18n"
)

type (
	// Context wraps Request and Response. It also provides methods for handling responses.
	Context struct {
		Request  *Request
		Response ResponseWriter
		Router   *Router
		Params   map[string]string
		meta     map[string]string
		Session  Session
		I18n     *i18n.I18N
		Shared   map[string]interface{}

		Data map[string]interface{}
	}
)

// NewContext creates new context for the given w and r.
func NewContext(w ResponseWriter, r *Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
		Data:     make(map[string]interface{}),
		Params:   make(map[string]string),
		meta:     make(map[string]string),
		Shared:   make(map[string]interface{}),
	}
}

// Reset the context information.
func (c *Context) Reset() {
	c.Data = make(map[string]interface{})
	c.Params = make(map[string]string)
	c.meta = make(map[string]string)
}

// Header sets or deletes the response headers.
func (c *Context) Header(key, value string) {
	if value == "" {
		c.Response.Header().Del(key)
	} else {
		c.Response.Header().Set(key, value)
	}
}

// Plain sends a string.
func (c *Context) Plain(code int, s string) {
	c.Response.WriteHeader(code)
	_, _ = c.Response.Write([]byte(s))
}

// JSON renders as application/json with the given code.
func (c *Context) JSON(code int, obj interface{}) {
	header := c.Response.Header()
	header["Content-Type"] = []string{"application/json; charset=utf-8"}

	c.Response.WriteHeader(code)
	_ = json.NewEncoder(c.Response).Encode(obj)
}

// Redirect the request to other url location.
func (c *Context) Redirect(status int, url string) {
	http.Redirect(c.Response, c.Request.Request, url, status)
}

// Render outputs using specified template.
func (c *Context) Render(name string) {
	if len(c.meta) > 0 {
		c.Data["FrameworkMeta"] = c.meta
	}

	c.Data["Request"] = c.Request
	c.Data["i18n"] = c.I18n

	if len(App.sharedData) > 0 {
		for key, value := range App.sharedData {
			c.Data[key] = value
		}
	}

	_ = App.View.Render(c.Response, name, c.Data)
}

// Meta sets or deletes the meta tags.
func (c *Context) Meta(key, value string) {
	if value == "" {
		delete(c.meta, key)
	} else {
		c.meta[key] = value
	}
}

// ParseForm detects the post Content-Type and prepares the form fields.
// If the Content-Type is application/json the values are decoded into the request JSON and JSONRaw params.
func (c *Context) ParseForm() error {
	c.Request.JSON = &JSONData{}
	c.Request.JSONRaw = ""

	_ = c.Request.ParseForm()

	if strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
		b, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			return fmt.Errorf("error reading body: %v\n", err)
		}

		if len(b) > 0 {
			parsedJSON, err := PrepareJSON(b)
			if err != nil {
				_ = c.Request.Body.Close()
				return fmt.Errorf("invalid JSON: %v\n", err)
			}

			_ = c.Request.Body.Close()
			c.Request.JSON = parsedJSON
			c.Request.JSONRaw = string(b)
		}
	}

	return nil
}

// ParseJSON parse the body into your defined structure.
func (c *Context) ParseJSON(v interface{}) error {
	b, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %v\n", err)
	}

	err = json.Unmarshal(b, &v)
	if err != nil {
		_ = c.Request.Body.Close()
		return fmt.Errorf("invalid JSON: %v\n", err)
	}

	return nil
}

// Error manage the error based on code
func (c *Context) Error(code int, err error) {
	// In case of error 500, log as critical and print stack trace
	if code == http.StatusInternalServerError {
		Log.Critical(fmt.Sprintf("%v", err))
		debug.PrintStack()
	} else {
		Log.Error(err)
	}

	c.Plain(code, err.Error())
}

// AddShared append value to context shared values
func (c *Context) AddShared(key string, value interface{}) {
	c.Shared[key] = value
}

// GetShared get value from context shared values
func (c *Context) GetShared(key string) interface{} {
	return c.Shared[key]
}

// ParseUploads
