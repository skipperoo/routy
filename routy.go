// Package routy provides a lightweight HTTP router built on top of the standard
// net/http ServeMux. It adds support for middlewares, subroutes, and handler
// registration in a chainable, minimal API.
package routy

import (
	"net/http"
	"strings"
)

// handler represents a single HTTP endpoint handler registered to the router.
type handler struct {
	Endpoint string
	fn       http.HandlerFunc
}

// middleware defines a function that wraps an http.Handler, allowing
// pre-processing or post-processing of requests (e.g., logging, auth).
type (
	middleware func(http.Handler) http.Handler
	subroute   struct {
		Path   string
		Handle http.Handler
	}
)

// Router is the core structure of the routy package.
// It wraps an http.ServeMux and maintains lists of handlers, subroutes,
// and middlewares that can be composed before serving.
type Router struct {
	r           *http.ServeMux
	handlers    []handler
	middlewares []middleware
	subroutes   []subroute
}

// NewRouter creates and returns a new instance of Router
// with an initialized http.ServeMux.
func NewRouter() *Router {
	router := http.NewServeMux()

	return &Router{r: router}
}

// AddHandler registers a new HTTP handler function for the given endpoint path.
// It returns the router instance to allow method chaining.
//
// Example:
//
//	r.AddHandler("/ping", func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprintln(w, "pong")
//	})
func (r *Router) AddHandler(endpoint string, fn http.HandlerFunc) *Router {
	r.handlers = append(r.handlers, handler{endpoint, fn})
	return r
}

// AddMiddleware registers a middleware function that wraps the router's handlers.
// Middlewares are executed in the order they are added (outermost to innermost).
//
// Example:
//
//	r.AddMiddleware(loggingMiddleware)
func (r *Router) AddMiddleware(mw middleware) *Router {
	r.middlewares = append(r.middlewares, mw)
	return r
}

// AddSubroute registers a subroute (a nested router or handler) under a given path prefix.
// The handler for the subroute will receive requests with the prefix stripped.
//
// Example:
//
//	apiRouter := routy.NewRouter()
//	r.AddSubroute("/api/", apiRouter.Finalize())
func (r *Router) AddSubroute(path string, handler http.Handler) *Router {
	r.subroutes = append(r.subroutes, subroute{path, handler})
	return r
}

// createStack combines a list of middleware functions into a single middleware
// that wraps an http.Handler in reverse order of registration.
func createStack(mw []middleware) middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			x := mw[i]
			next = x(next)
		}
		return next
	}
}

// Finalize finalizes the router configuration and returns the composed http.Handler.
// It registers all handlers, subroutes, and wraps the ServeMux with the middleware stack.
//
// Example:
//
//	http.ListenAndServe(":8080", r.Finalize())
func (r *Router) Finalize() http.Handler {
	for _, handler := range r.handlers {
		r.r.HandleFunc(handler.Endpoint, handler.fn)
	}
	for _, subroute := range r.subroutes {
		prefix := strings.TrimSuffix(subroute.Path, "/")
		strippedHandler := http.StripPrefix(prefix, subroute.Handle)

		r.r.Handle(subroute.Path, strippedHandler)
	}
	stack := createStack(r.middlewares)

	return stack(r.r)
}
