// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AnUnnamedProject/framework/utils"
)

type (
	// Session contains the base session provider.
	// Load configuration automatically from app.json file
	Session interface {
		// Init checks if session id exists and return it.
		// if session does not exists, creates a new one.
		Init(http.ResponseWriter, *http.Request) (Session, error)
		// Release writes back the session header
		Release(http.ResponseWriter, *http.Request)
		// Set sets value to given key in session.
		Set(key string, value interface{}) error
		// Get gets value by given key in session.
		Get(string) interface{}
		// Delete deletes a key from session.
		Delete(string) error
		// Destroy the entire session
		Destroy() error
		// SessionID returns current session ID.
		SessionID() string
		// Handle
		Handler(c *Context)
	}

	// SessionProvider contains the session instance.
	SessionProvider func(config string) (Session, error)

	// SessionConfig contains the session instance default configuration.
	SessionConfig struct {
		// Name is the cookie session name
		Name string `json:"name"`
		// //
		// Config string `json:"config"`
		// Domain limits cookie access to specified domain
		Domain string `json:"domain"`
		// MaxLifetime as seconds
		MaxLifetime int `json:"max_lifetime"`
		// Key is a private key used to encrypt session cookies
		Key string `json:"key"`
		// HTTPOnly allow access to cookies only via HTTP requests,
		// Set to true to protect against XSS exploits
		HTTPOnly bool `json:"http_only"`
		// Hash     hash.Hash
	}
)

var sessionProviders = make(map[string]Session)

// NewSession creates the session provider instance using the provided name as provider
// and add it to the available session providers.
func NewSession(name string, config string, provider SessionProvider) (Session, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	if _, ok := sessionProviders[name]; ok {
		return sessionProviders[name], nil
	}

	session, err := provider(config)
	sessionProviders[name] = session

	return session, err
}

// EncodeCookie return a new encrypted cookie string.
func EncodeCookie(key []byte, sessionID string, req *http.Request) string {
	ts := time.Now().Unix()

	value := &bytes.Buffer{}
	_, _ = value.WriteString(sessionID)
	_, _ = value.WriteString("|")
	_, _ = value.WriteString(strconv.FormatInt(ts, 10))
	_, _ = value.WriteString("|")
	_, _ = value.WriteString(createSignature(req))

	return utils.EncryptXOR(key, value.Bytes())
}

// DecodeCookie decrypt and return the cookie id.
func DecodeCookie(key []byte, maxlifetime int, signedValue string, req *http.Request) (string, error) {
	if signedValue == "" {
		return "", fmt.Errorf("value is empty")
	}

	// Decode Base64 cookie and decrypt using provided Key
	decrypted := utils.DecryptXOR(key, signedValue)

	// Split
	parts := strings.Split(string(decrypted), "|")

	if len(parts) != 3 {
		return "", fmt.Errorf("invalid signed value")
	}

	sessionID := strings.Trim(parts[0], "\x00")
	ts := parts[1]
	s := parts[2]

	// Check Signature
	newSignature := createSignature(req)
	if s != newSignature {
		return sessionID, fmt.Errorf("invalid signature")
	}

	// Check Timestamp (expired, tampering)
	var t int64
	t, err := strconv.ParseInt(ts, 0, 64)
	if err != nil {
		return sessionID, fmt.Errorf("invalid timestamp")
	}

	cookieTime := time.Unix(t, 0)

	// Expired
	cookieTime.Before(time.Now().AddDate(0, 0, -31))

	// Future
	if cookieTime.After(time.Now()) {
		return sessionID, fmt.Errorf("session date from future")
	}

	return sessionID, nil
}

func createSignature(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return ""
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return ""
	}

	// This will only be defined when site is accessed via non-anonymous proxy
	// and takes precedence over RemoteAddr
	// Header.Get is case-insensitive
	forward := req.Header.Get("X-Forwarded-For")

	if forward != "" {
		ip = forward
	}

	return ip + "/" + req.Header.Get("user-agent")
}

// GetSessionID return the cookie from http request.
func GetSessionID(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return ""
	}

	cookie.Value, err = url.QueryUnescape(cookie.Value)
	if err != nil {
		return ""
	}

	return cookie.Value
}

// NewSessionID return a new random cookie.
func NewSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
