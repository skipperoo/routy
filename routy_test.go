package routy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Helpers ---

func makeRequest(t *testing.T, handler http.Handler, method, path string) string {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	return string(body)
}

// --- Tests ---

func TestAddHandler(t *testing.T) {
	r := NewRouter()
	r.AddHandler("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world"))
	})

	handler := r.Finalize()

	body := makeRequest(t, handler, http.MethodGet, "/hello")
	if body != "world" {
		t.Fatalf("expected body 'world', got %q", body)
	}
}

func TestPathParameter(t *testing.T) {
	r := NewRouter()
	r.AddHandler("/hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		w.Write([]byte(name))
	})

	handler := r.Finalize()

	body := makeRequest(t, handler, http.MethodGet, "/hello/routy")
	if body != "routy" {
		t.Fatalf("expected body 'routy', got %q", body)
	}
}

func TestMiddlewareOrder(t *testing.T) {
	var order []string

	// middleware1 wraps middleware2 wraps handler
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1-before")
			next.ServeHTTP(w, r)
			order = append(order, "m1-after")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2-before")
			next.ServeHTTP(w, r)
			order = append(order, "m2-after")
		})
	}

	r := NewRouter().
		AddMiddleware(m1).
		AddMiddleware(m2).
		AddHandler("/test", func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

	handler := r.Finalize()

	makeRequest(t, handler, http.MethodGet, "/test")

	expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if strings.Join(order, ",") != strings.Join(expected, ",") {
		t.Fatalf("unexpected middleware order:\n got  %v\n want %v", order, expected)
	}
}

func TestSubroute(t *testing.T) {
	api := NewRouter().
		AddHandler("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong"))
		})

	r := NewRouter().
		AddSubroute("/api/", api.Finalize())

	handler := r.Finalize()

	body := makeRequest(t, handler, http.MethodGet, "/api/ping")
	if body != "pong" {
		t.Fatalf("expected 'pong', got %q", body)
	}
}

func TestChainedConfiguration(t *testing.T) {
	handler := NewRouter().
		AddHandler("/foo", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("bar"))
		}).
		AddMiddleware(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "ok")
				next.ServeHTTP(w, r)
			})
		}).Finalize()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Test") != "ok" {
		t.Fatalf("expected middleware to set X-Test header")
	}
	if body := rec.Body.String(); body != "bar" {
		t.Fatalf("expected body 'bar', got %q", body)
	}
}
