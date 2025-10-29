# routy

`routy` is a lightweight, composable HTTP router built on top of Go’s standard [`net/http`](https://pkg.go.dev/net/http) package.  
It allows you to register handlers, subroutes, and middlewares with a clean, fluent API.

---

##  Installation

```bash
go get github.com/skipperoo/routy
```

## Basic Usage

Here's an example project structure:
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

To create a basic router:

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/skipperoo/routy"
)

func main() {
	router := routy.NewRouter()

	router.AddHandler("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})
    router.Finalize()

	http.ListenAndServe(":8080", router)
}

```

You can also specify methods and use path parameters:

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
You can add middlewares and combine them with subroutes:
```go

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		// Validate your token
    })
}

func main() {

	router := routy.NewRouter()

	router.AddHandler("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	}).
    AddMiddleware(Logging)
    subRouter := routy.NewRouter().
    AddHandler("/supersecret", func(w http.ResponseWriter, r *http.Request){}).
    AddMiddleware(Auth).
    Finalize()

    // Now /hello will be public, while /auth/* will be protected,
    // while both will use the Logging middleware
    router.AddSubroute("/auth/", subRoute)
    router.Finalize()

	http.ListenAndServe(":8080", router)
}

```
Routy also provides two default configurable middlewares:
- `LoggingMiddleware` to log requests. It accepts a function to write the log, which by default is set to `log.Printf`.
- `RecoverMiddleware` to recover from panics while serving a request.
```go
type RecoverFunction func(w http.ResponseWriter, r *http.Request)
recoverMw := routy.NewRecoverMiddleware(nil)

type LoggingFunction func(format string, v ...any)
func logToFile(format string, a ...any) {
	f, err := os.OpenFile(os.Getenv("LOG_FILE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, format, a...)
}
loggingMw := routy.NewLoggingMiddleware(logToFile)

router.
	AddMiddleware(recoverMw.GetMiddleware()).
	AddMiddleware(loggingMw.GetMiddleware()).
	AddHandler("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})
```


