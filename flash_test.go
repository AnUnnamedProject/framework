// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlash(t *testing.T) {
	config := `{"name":"sessionid","max_lifetime":3600,"key":"somerandomkey234","save_path":"/tmp"}`

	session, err := NewSession("file", config, NewFileSessionProvider)
	if err != nil {
		t.Fatal("unable to create session")
	}

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c := NewContext(NewResponseWriter(w), &Request{Request: r})

	c.Session, err = session.Init(c.Response, c.Request.Request)
	if err != nil || c.Session == nil {
		t.Fatal("unable to start session", err)
	}

	// Write back session id to client
	c.Response.Before(func(bw ResponseWriter) {
		c.Session.Release(bw, c.Request.Request)
	})

	a := NewFlash()
	a.Set("success", "This is a success message!")

	if err = c.WriteFlash(a); err != nil {
		t.Fatal("unable to save flash", err)
	}

	values, err := c.GetFlash()
	if err != nil {
		t.Fatal("unable to unmarshal flash message", err)
	}

	if values == nil || values["success"] != "This is a success message!" {
		t.Fatal("unable to get flash message")
	}
}
