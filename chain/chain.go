package chain

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
)

type contextKey int

const (
	contextKeyChainIndex contextKey = iota
)

// A Config defines optional configs for running chain handlers
type Config struct{}

// The HandlerFunc type is an adapter to pass key-value pair from URL path with
// regular http handler w and r.
type HandlerFunc func(k, v string, w http.ResponseWriter, r *http.Request)

// Handler is request handler and chain executer.
// A pair of path segments is a key-value set for a chain handler.
// "/foo/bar" is handled by a handler defined for "foo"
type Handler struct {
	mu sync.RWMutex
	m  map[string]HandlerFunc
}

// NewHandler returns chain Handler with the default handler map.
func NewHandler(cfg *Config) *Handler {
	m := make(map[string]HandlerFunc)
	if cfg == nil {
		for name, fn := range defaultHandlerFuncs {
			m[name] = fn
		}
	}
	return &Handler{m: m}
}

// SetChainHandlerFunc defines HandlerFunc for a path segment.
func (hdr *Handler) SetChainHandlerFunc(seg string, fn HandlerFunc) {
	hdr.mu.Lock()
	defer hdr.mu.Unlock()

	// wrap fn with the chain logic
	hdr.m[seg] = func(k, v string, w http.ResponseWriter, r *http.Request) {
		fn(k, v, w, r)

		i := r.Context().Value(contextKeyChainIndex).(int)
		// skip the first ""(1), the current pair(2) and "i" number of pairs
		idx := 1 + 2 + i*2
		items := strings.Split(r.URL.Path, "/")[idx:]
		if len(items) > 0 {
			if nextFn, ok := hdr.ChainHandlerFunc(items[0]); ok {
				// increment chain index
				r = r.WithContext(context.WithValue(r.Context(), contextKeyChainIndex, i+1))
				log.Println("next", i+1, items[0], items[1])
				nextFn(items[0], items[1], w, r)
			}
		}
	}
}

// ChainHandlerFunc returns HandlerFunc for a path segment.
func (hdr *Handler) ChainHandlerFunc(path string) (HandlerFunc, bool) {
	hdr.mu.RLock()
	defer hdr.mu.RUnlock()

	fn, ok := hdr.m[path]
	return fn, ok
}

// ServeHTTP implements http.Handler interface. It finds HandlerFunc for the
// first path segment and start chaining handlers.
func (hdr *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// NOTE: Split("/", "/") returns ["", ""]
	items := strings.Split(r.URL.Path, "/")[1:]

	if len(items[0]) > 0 { // check key string length
		if fn, ok := hdr.ChainHandlerFunc(items[0]); ok {
			// initialize chain index to 0
			r = r.WithContext(context.WithValue(r.Context(), contextKeyChainIndex, 0))
			fn(items[0], items[1], w, r)
			return
		}
	}

	rootHandler(w, r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
