package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	listen = flag.String("listen", ":8128", "TCP address `host:port` to listen on")
)

type contextKey int

const (
	contextKeyChainIndex contextKey = iota
)

func init() {
	flag.Parse()
}

type chain struct {
	m map[string]http.HandlerFunc
}

func (c *chain) Add(path string, hdr http.HandlerFunc) {
	c.m[path] = hdr
}

func (c *chain) HandlerFunc(path string) (http.HandlerFunc, bool) {
	fn, ok := c.m[path]
	return fn, ok
}

func (c *chain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, fn := range c.m {
		if strings.HasPrefix(r.URL.Path, "/"+k+"/") {
			// initialize chain index
			r = r.WithContext(context.WithValue(r.Context(), contextKeyChainIndex, 0))
			fn(w, r)
			return
		}
	}
	rootHandler(w, r)
}

var c = &chain{make(map[string]http.HandlerFunc)}

func main() {
	c.Add("status", statusHandler())
	c.Add("delay", delayHandler())

	log.Fatal(http.ListenAndServe(*listen, c))
}

func chainWrapper(fn func(string, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		i := r.Context().Value(contextKeyChainIndex).(int)
		idx := 2 + i*2 // skip 2 (current key-value pair) and "i" number of pairs
		items := strings.Split(r.URL.Path, "/")[idx:]
		if len(items) > 0 {
			fn(items[0], w, r)
			if len(items) > 1 {
				if next, ok := c.HandlerFunc(items[1]); ok {
					r = r.WithContext(context.WithValue(r.Context(), contextKeyChainIndex, i+1))
					next(w, r)
				}
			}
		}
	}
}

func statusHandler() func(http.ResponseWriter, *http.Request) {
	return chainWrapper(func(s string, w http.ResponseWriter, r *http.Request) {
		statusCode, err := strconv.Atoi(s)
		if err != nil {
			log.Println(err)
			statusCode = http.StatusOK
		}
		log.Printf("stauts %d", statusCode)
		w.WriteHeader(int(statusCode))
	})
}

func delayHandler() func(http.ResponseWriter, *http.Request) {
	return chainWrapper(func(s string, w http.ResponseWriter, r *http.Request) {
		sleep, err := strconv.ParseInt(s, 10, 0) // to get int64
		if err != nil {
			log.Println(err)
			sleep = int64(0)
		}
		d := time.Duration(sleep * int64(time.Second))
		log.Printf("delay %v", d)
		time.Sleep(d)
	})

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
