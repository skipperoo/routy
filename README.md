# 🛣️ routy

`routy` is a lightweight, composable HTTP router built on top of Go’s standard [`net/http`](https://pkg.go.dev/net/http) package.  
It allows you to register handlers, subroutes, and middlewares with a clean, fluent API.

---

## 📦 Installation

```bash
go get github.com/yourusername/routy
```

## Basic Usage

```
.
├── main.go
├── internal/
│   ├── handler/
│   │   └── handler1.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── cors.go
│   └── service/
│       └── func.go
└── go.mod
```

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/yourusername/routy"
)

func main() {
	router := routy.NewRouter()

	router.AddHandler("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	http.ListenAndServe(":8080", router.Finalize())
}

```

### Routes
```go
router.
	AddHandler("GET /users", getUsers).
	AddHandler("POST /users", createUser).
	AddHandler("GET /users/{id}", getUserByID)

func getUsers(w http.ResponseWriter, r* http.Request) {
    id := r.PathValue("id")
    // Use id
	fmt.Fprintf(w, "User ID: %s", id)
}
```

### Middlewares
#### Built-in
```go
recoverMw := routy.NewRecoverMiddleware(nil)
loggingMw := routy.NewLoggingMiddleware(nil)

router.
	AddMiddleware(recoverMw.GetMiddleware()).
	AddMiddleware(loggingMw.GetMiddleware()).
	AddHandler("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})
```

#### Custom middlewares
```go

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		// Validate your token
    })
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

```
