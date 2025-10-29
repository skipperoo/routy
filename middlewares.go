package routy

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type (
	// RecoverFunction defines the function signature used by RecoverMiddleware
	// to handle a panic recovery. It is invoked in a deferred context whenever
	// a panic occurs inside a wrapped handler.
	//
	// The function is responsible for writing an appropriate response.
	RecoverFunction func(w http.ResponseWriter, r *http.Request)

	// LoggingFunction defines the function signature used by LoggingMiddleware
	// for logging requests. It typically accepts a format string followed by
	// variadic arguments, similar to log.Printf.
	LoggingFunction func(format string, v ...any)

	// RecoverMiddleware provides panic recovery for HTTP handlers.
	// If a panic occurs, it executes the provided RecoverFunction.
	RecoverMiddleware struct {
		recoverFunction RecoverFunction
	}

	// LoggingMiddleware logs basic information about each HTTP request, including
	// status code, method, path, and duration. It uses a user-provided or default
	// logging function to output log messages.
	LoggingMiddleware struct {
		loggingFunction LoggingFunction
	}

	// IMiddleware represents a generic middleware component capable of producing
	// a routy-compatible middleware function.
	IMiddleware interface {
		GetMiddleware() middleware
	}
)

// wrappedWriter wraps an http.ResponseWriter to capture the status code
// written by handlers. It defaults to 200 OK if no explicit status is written.
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader records the HTTP status code before delegating to the underlying
// ResponseWriter.
func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

// NewRecoverMiddleware creates a new RecoverMiddleware. If a nil recoverFunction
// is provided, it assigns a default one that logs the panic, prints a stack trace,
// and returns a generic 500 Internal Server Error response.
//
// Example:
//
//	rm := routy.NewRecoverMiddleware(nil)
//	router.AddMiddleware(rm.GetMiddleware())
func NewRecoverMiddleware(recoverFunction RecoverFunction) *RecoverMiddleware {
	if recoverFunction == nil {
		recoverFunction = func(w http.ResponseWriter, r *http.Request) {
			if err := recover(); err != nil {
				log.Printf("Caught panic: %v. Stack trace: %s\n", err, string(debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Internal server error")
			}
		}
	}
	return &RecoverMiddleware{
		recoverFunction: recoverFunction,
	}
}

// GetMiddleware returns a routy-compatible middleware function that wraps
// the provided handler in a panic recovery mechanism.
//
// Example:
//
//	rm := routy.NewRecoverMiddleware(nil)
//	router.AddMiddleware(rm.GetMiddleware())
func (rm *RecoverMiddleware) GetMiddleware() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer rm.recoverFunction(w, r)
			next.ServeHTTP(w, r)
		})
	}
}

// NewLoggingMiddleware creates a new LoggingMiddleware. If a nil loggingFunction
// is provided, it defaults to using log.Printf.
//
// Example:
//
//	lm := routy.NewLoggingMiddleware(nil)
//	router.AddMiddleware(lm.GetMiddleware())
func NewLoggingMiddleware(loggingFunction LoggingFunction) *LoggingMiddleware {
	if loggingFunction == nil {
		loggingFunction = log.Printf
	}
	return &LoggingMiddleware{
		loggingFunction: loggingFunction,
	}
}

// GetMiddleware returns a routy-compatible middleware function that logs each
// HTTP request. It records the status code, method, path, and request duration.
//
// Example:
//
//	lm := routy.NewLoggingMiddleware(nil)
//	router.AddMiddleware(lm.GetMiddleware())
func (lm *LoggingMiddleware) GetMiddleware() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &wrappedWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(wrapped, r)
			lm.loggingFunction("%v %v %v %v", wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))
		})
	}
}
