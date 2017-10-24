// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"errors"
	"net/http"
)

// Controller interface for controller.
type Controller interface {
	Route(pattern string, method string, handler HandlerFunc)
	GET(pattern string, handlers ...HandlerFunc)
	POST(pattern string, handlers ...HandlerFunc)
	PUT(pattern string, handlers ...HandlerFunc)
	DELETE(pattern string, handlers ...HandlerFunc)
	PATCH(pattern string, handlers ...HandlerFunc)
	OPTIONS(pattern string, handlers ...HandlerFunc)
	HEAD(pattern string, handlers ...HandlerFunc)
	Any(pattern string, handlers ...HandlerFunc)
}

// BaseController implements the Controller.
// All user defined controllers must use BaseController.
type BaseController struct {
	Router   Router
	Layout   string
	Meta     map[string]string
	Response http.ResponseWriter
	Request  *http.Request
}

// Route define a route for the current controller.
// usage:
//	default methods is the same name as method
//	Route("/user", "GET", UserController.Get)
//	Route("/api/list", "GET", ApiController.List)
//	Route("/api/create", "POST", ApiController.Save)
//	Route("/api/update", "PUT", ApiController.Save)
//	Route("/api/delete", "DELETE", ApiController.Delete)
func (c *BaseController) Route(pattern string, method string, handler HandlerFunc) {
	if pattern == "" {
		Log.Error(errors.New("please enter a valid pattern"))
	}

	if pattern[0] != '/' {
		Log.Error(errors.New(`path must begin with "/"`))
	}

	App.Router.Add(pattern, method, handler)
}

// GET is an alias to Add(pattern, "GET", handlers)
func (c *BaseController) GET(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "GET", handler)
	}
}

// POST is an alias to Add(pattern, "POST", handlers)
func (c *BaseController) POST(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "POST", handler)
	}
}

// PUT is an alias to Add(pattern, "PUT", handlers)
func (c *BaseController) PUT(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "PUT", handler)
	}
}

// DELETE is an alias to Add(pattern, "DELETE", handlers)
func (c *BaseController) DELETE(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "DELETE", handler)
	}
}

// PATCH is an alias to Add(pattern, "PATCH", handlers)
func (c *BaseController) PATCH(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "PATCH", handler)
	}
}

// OPTIONS is an alias to Add(pattern, "OPTIONS", handlers)
func (c *BaseController) OPTIONS(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "OPTIONS", handler)
	}
}

// HEAD is an alias to Add(pattern, "HEAD", handlers)
func (c *BaseController) HEAD(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "HEAD", handler)
	}
}

// Any register a route that matches all HTTP methods.
// (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
func (c *BaseController) Any(pattern string, handlers ...HandlerFunc) {
	for _, handler := range handlers {
		c.Route(pattern, "GET", handler)
		c.Route(pattern, "POST", handler)
		c.Route(pattern, "PUT", handler)
		c.Route(pattern, "DELETE", handler)
		c.Route(pattern, "PATCH", handler)
		c.Route(pattern, "OPTIONS", handler)
		c.Route(pattern, "HEAD", handler)
	}
}

// RegisterController register the specified controller on framework.
// controller func Init is called on framework initialization.
func RegisterController(c Controller) {
	App.Controllers = append(App.Controllers, c)
}
