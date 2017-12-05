// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AnUnnamedProject/i18n"
)

type (
	// Router is the framework router interface
	Router interface {
		Add(pattern string, method string, handlers ...HandlerFunc)
		ServeHTTP(rw http.ResponseWriter, req *http.Request)
	}

	// Routes contains the framework routes, middleware and custom http error messages
	Routes struct {
		routes     []*Route
		httpErrors map[int]HandlerFunc
	}

	// Route contain the single route structure
	Route struct {
		method   string
		regex    *regexp.Regexp
		params   map[int]string
		handlers []HandlerFunc
	}

	// HandlerFunc defines the function type for controller requests.
	HandlerFunc func(c *Context)
)

// NewRouter instantiate a new Router
func NewRouter() Router {
	r := &Routes{}
	r.httpErrors = make(map[int]HandlerFunc)
	return r
}

// Add adds a new route to the app routes.
func (r *Routes) Add(pattern string, method string, handlers ...HandlerFunc) {
	if pattern == "" {
		Log.Error(errors.New("please enter a valid pattern"))
	}

	if pattern[0] != '/' {
		Log.Error(errors.New(`path must begin with "/"`))
	}

	if method == "" {
		Log.Error(errors.New("please enter a valid method"))
	}

	parts := strings.Split(pattern, "/")

	// Update dynamic parts that contains a : with a regexp
	j := 0
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	// Rejoin parts to the pattern
	pattern = strings.Join(parts, "/")
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		panic(regexErr)
	}

	// handlers are inverted, first handler is the latest in execution order
	// i.e. Execute first all middlewares and then the controller handler
	reverseHandlers := make([]HandlerFunc, len(handlers))
	for i := len(handlers) - 1; i >= 0; i-- {
		reverseHandlers[i] = handlers[i]
	}

	route := &Route{}
	route.method = method
	route.regex = regex
	route.handlers = reverseHandlers
	route.params = params

	if Mode() == DebugMode {
		Log.Info(fmt.Sprintf("Adding route [%s] %s", method, pattern))
	}

	r.routes = append(r.routes, route)
}

// ServeHTTP handle the request based on defined routes.
func (r *Routes) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c := App.pool.Get(NewResponseWriter(rw), &Request{Request: req})
	defer App.pool.Put(c)

	// Recover
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	path := req.URL.Path
	routeFound := false

	language := c.Request.Header.Get("Accept-Language")
	if language != "" && strings.Contains(language, ",") {
		language = language[:strings.Index(language, ",")]
	}
	c.I18n = i18n.New(language)

	// Execute global middlewares
	for _, filter := range App.middlewares {
		filter(c)

		// Manage middleware error (f.e. BasicAuth error)
		if c.Response.Status() != 0 {
			return
		}
	}

	// Check if requested path is a static file.
	servedStatic := r.ServeStaticFiles(c)
	if servedStatic {
		return
	}

	// Check if route exists
	for _, route := range r.routes {
		// if the method don't match, skip it
		if req.Method != route.method {
			continue
		}

		// if path regex doesn't match, skip it
		if !route.regex.MatchString(path) {
			continue
		}

		matches := route.regex.FindAllStringSubmatch(path, -1)

		// if path length is different from matches[0] length, skip it
		if len(matches[0][0]) != len(path) {
			continue
		}

		// process route params and rewrite the raw query to have them available
		// on request.
		if len(route.params) > 0 {
			values := req.URL.Query()
			for _, match := range matches {
				for j, param := range route.params {
					values.Add(param, match[j+1])
				}
			}

			req.URL.RawQuery = url.Values(values).Encode() + "&" + req.URL.RawQuery
		}

		// Route found, invoke handler(s)
		for _, handler := range route.handlers {
			handler(c)

			// Manage http errors
			if c.Response.Status() != 0 {
				return
			}
		}
		routeFound = true
	}

	// Route not found
	if !routeFound {
		c.Error(http.StatusNotFound, fmt.Errorf("404 page %s not found", path))
	}
}

// ServeStaticFiles check if requested path is a static file and serve the content.
func (r *Routes) ServeStaticFiles(c *Context) bool {
	if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
		return false
	}

	req := c.Request.URL.Path

	// Cache is configured
	if App.cache != nil {
		// If requested file is CSS and compress_css is enabled, minify and serve a cached version.
		// If the file is already minified (.min.css) we don`t perform any additional compression.
		if path.Ext(req) == ".css" && Config.Bool("compress_css") && !strings.Contains(req, ".min.css") {
			var css string
			req = "file:" + req

			// If file is not already cached: read, minify and put in cache
			if !App.cache.Exists(req) {
				filePath, fileInfo, _ := lookupFile(c.Request.URL.Path)
				if fileInfo == nil {
					// TODO: Logger should log this as an error
					return false
				}
				data, err := ioutil.ReadFile(filePath)
				if err != nil {
					// TODO: Logger should log this as an error
					return false
				}
				css = MinifyCSS(data)

				// Store in cache for one day
				_ = App.cache.Put(req, css, 3600*24*time.Second)
			} else {
				css = App.cache.Get(req).(string)
			}

			r := strings.NewReader(css)

			// TODO: Define per resource cache
			c.Header("Content-Length", fmt.Sprintf("%d", len(css)))
			c.Header("Cache-Control", "public, max-age=11111111")
			http.ServeContent(c.Response, c.Request.Request, c.Request.URL.Path, time.Now(), r)
			return true
		}

		// If requested file is JS and compress_js is enabled, minify and serve a cached version.
		// If the file is already minified (.min.js) we don`t perform any additional compression.
		if path.Ext(req) == ".js" && Config.Bool("compress_js") && !strings.Contains(req, ".min.js") {
			var js string
			req = "file:" + req

			// If file is not already cached: read, minify and put in cache
			if !App.cache.Exists(req) {
				filePath, fileInfo, _ := lookupFile(c.Request.URL.Path)
				if fileInfo == nil {
					// TODO: Logger should log this as an error
					return false
				}
				data, err := ioutil.ReadFile(filePath)
				if err != nil {
					// TODO: Logger should log this as an error
					return false
				}
				js = MinifyJS(data)

				// Store in cache for one day
				_ = App.cache.Put(req, js, 3600*24*time.Second)
			} else {
				js = App.cache.Get(req).(string)
			}

			r := strings.NewReader(js)

			// TODO: Define per resource cache
			c.Header("Content-Length", fmt.Sprintf("%d", len(js)))
			c.Header("Cache-Control", "public, max-age=11111111")
			http.ServeContent(c.Response, c.Request.Request, c.Request.URL.Path, time.Now(), r)
			return true
		}
	}

	// Check if file exists locally
	filePath, fileInfo, _ := lookupFile(req)
	if fileInfo == nil {
		return false
	}

	// If it's a directory, we don't provide directory listing
	if fileInfo.IsDir() {
		return false
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}

	// TODO: Caching per resource
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Cache-Control", "public, max-age=11111111")
	http.ServeContent(c.Response, c.Request.Request, filePath, fileInfo.ModTime(), file)
	_ = file.Close()
	return true
}

// OnError add a custom error handler
func (r *Routes) OnError(code int, handlerFunc HandlerFunc) {
	r.httpErrors[code] = handlerFunc
}
