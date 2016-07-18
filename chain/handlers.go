package chain

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

var defaultHandlerFuncs map[string]HandlerFunc

func init() {
	defaultHandlerFuncs = make(map[string]HandlerFunc)
	defaultHandlerFuncs["status"] = statusHandler
	defaultHandlerFuncs["delay"] = delayHandler
}

// /status/<status code>
//
// set http response status code. Unfortunately this chain handler needs to be
// the first one.
//
// example: /status/206
//
func statusHandler(k, v string, w http.ResponseWriter, r *http.Request) {
	statusCode, err := strconv.Atoi(v)
	if err != nil {
		log.Println(err)
		statusCode = http.StatusOK
	}
	w.WriteHeader(int(statusCode))
}

// /delay/<seconds>
//
// sleep da mount of time defined as `seconds` value.
//
// example: /delay/1
//
func delayHandler(k, v string, w http.ResponseWriter, r *http.Request) {
	sleep, err := strconv.ParseInt(v, 10, 0) // to get int64
	if err != nil {
		log.Println(err)
		sleep = int64(0)
	}
	d := time.Duration(sleep * int64(time.Second))
	time.Sleep(d)
}
