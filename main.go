package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/sekimura/forsok/chain"
)

var (
	listen = flag.String("listen", ":8128", "TCP address `host:port` to listen on")
)

func init() {
	flag.Parse()
}

func main() {
	// path nil to chain.NewServeMux to enable default chain handlers
	hdr := chain.NewHandler(nil)

	// it's easy to define a custom chain handler
	hdr.SetChainHandlerFunc("hello", func(k, v string, w http.ResponseWriter, r *http.Request) {
		log.Println("custom chain handler:", k, v)
		fmt.Fprintf(w, "HELLO %s %s\n", k, v)
	})

	log.Fatal(http.ListenAndServe(*listen, hdr))
}
