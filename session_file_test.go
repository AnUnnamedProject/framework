// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFileSession(t *testing.T) {
	config := `{"name":"sessionid","max_lifetime":3600,"key":"somerandomkey234","save_path":"/tmp"}`

	provider, err := NewSession("file", config, NewFileSessionProvider)
	if err != nil {
		t.Fatal("init session", err)
	}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, err := provider.Init(w, r)
	if err != nil || sess == nil {
		t.Fatal("unable to start session", err)
	}
	defer sess.Release(w, r)

	err = sess.Set("user_id", "1")
	if err != nil {
		t.Fatal("set error,", err)
	}

	if userID := sess.Get("user_id"); userID != "1" {
		t.Fatal("get error")
	}

	if err := sess.Destroy(); err != nil {
		t.Fatal("unable to destroy session", err)
	}
}
