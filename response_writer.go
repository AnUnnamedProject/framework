// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type (
	// ResponseWriter extends http.ResponseWriter.
	ResponseWriter interface {
		http.ResponseWriter
		http.Hijacker

		// Before calls a function before the ResponseWriter has been written.
		Before(BeforeFunc)
		Status() int
	}
	responseWriter struct {
		http.ResponseWriter
		beforeFuncs []BeforeFunc
		status      int
	}
	// BeforeFunc defines a function called before the end of response.
	BeforeFunc func(ResponseWriter)
)

// NewResponseWriter creates a ResponseWriter that wraps an http.ResponseWriter
func NewResponseWriter(rw http.ResponseWriter) ResponseWriter {
	return &responseWriter{rw, nil, 0}
}

// WriteHeader writes a custom status code to header.
func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.callBefore()
	rw.ResponseWriter.WriteHeader(s)
}

// Before add a BeforeFunc to the functions called before the end of response.
func (rw *responseWriter) Before(before BeforeFunc) {
	rw.beforeFuncs = append(rw.beforeFuncs, before)
}

// Status returns current ResponseWriter status.
func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) callBefore() {
	for i := len(rw.beforeFuncs) - 1; i >= 0; i-- {
		rw.beforeFuncs[i](rw)
	}
}

// Hijack for http connection (required for websockets)
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("webserver doesn't support hijacking")
	}
	return hj.Hijack()
}
