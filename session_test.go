// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"net/http"
	"reflect"
	"testing"
)

func BenchmarkEncodeCookie(b *testing.B) {
	r, _ := http.NewRequest("GET", "/", nil)

	k := []byte("TEST")

	for i := 0; i < b.N; i++ {
		EncodeCookie(k, "TEST", r)
	}
}

func BenchmarkDecodeCookie(b *testing.B) {
	r, _ := http.NewRequest("GET", "/", nil)

	k := []byte("TEST")
	encoded := EncodeCookie(k, "TEST", r)

	for i := 0; i < b.N; i++ {
		_, _ = DecodeCookie(k, 1000, encoded, r)
	}
}

func BenchmarkSignature(b *testing.B) {
	r, _ := http.NewRequest("GET", "/", nil)

	for i := 0; i < b.N; i++ {
		createSignature(r)
	}
}

func TestNewSession(t *testing.T) {
	type args struct {
		name     string
		config   string
		provider SessionProvider
	}
	tests := []struct {
		name    string
		args    args
		want    Session
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := NewSession(tt.args.name, tt.args.config, tt.args.provider)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. NewSession() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. NewSession() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNewSessionID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := NewSessionID(); got != tt.want {
			t.Errorf("%q. NewSessionID() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
