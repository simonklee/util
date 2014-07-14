// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pat

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Creates a stack of HTTP handlers. Each HTTP handler is responsible for
// calling the next. The handlers are executed in reverse order, the last is
// called first.
func use(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

// registerPat a pattern with a handler for the given request method.
func registerPat(r *mux.Router, meth, pat string, h http.Handler, middleware []func(http.Handler) http.Handler) *mux.Route {
	return r.NewRoute().PathPrefix(pat).Handler(use(h, middleware...)).Methods(meth)
}

// Delete registers a pattern with a handler for DELETE requests.
func Delete(r *mux.Router, pat string, h http.Handler, middleware ...func(http.Handler) http.Handler) *mux.Route {
	return registerPat(r, "DELETE", pat, h, middleware)
}

// Get registers a pattern with a handler for GET requests.
func Get(r *mux.Router, pat string, h http.Handler, middleware ...func(http.Handler) http.Handler) *mux.Route {
	return registerPat(r, "GET", pat, h, middleware)
}

// Head registers a pattern with a handler for HEAD requests.
func Head(r *mux.Router, pat string, h http.Handler, middleware ...func(http.Handler) http.Handler) *mux.Route {
	return registerPat(r, "HEAD", pat, h, middleware)
}

// Options registers a pattern with a handler for OPTIONS requests.
func Options(r *mux.Router, pat string, h http.Handler, middleware ...func(http.Handler) http.Handler) *mux.Route {
	return registerPat(r, "OPTIONS", pat, h, middleware)
}

// Post registers a pattern with a handler for POST requests.
func Post(r *mux.Router, pat string, h http.Handler, middleware ...func(http.Handler) http.Handler) *mux.Route {
	return registerPat(r, "POST", pat, h, middleware)
}

// Put registers a pattern with a handler for PUT requests.
func Put(r *mux.Router, pat string, h http.Handler, middleware ...func(http.Handler) http.Handler) *mux.Route {
	return registerPat(r, "PUT", pat, h, middleware)
}
