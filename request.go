// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"net"
	"net/http"
	"strings"
)

type (
	// Request extends the default http.Request.
	Request struct {
		*http.Request

		JSON *JSONData
		// JsonRaw contains the raw json message.
		JSONRaw string
	}
)

func (r *Request) IP() string {
	// x-forwarded-for: client, proxy1, proxy2, ...
	if proxy := r.Header.Get("x-forwarded-for"); proxy != "" {
		proxy = strings.Split(proxy, ",")[0]
		if proxy != "" {
			return proxy
		}
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
