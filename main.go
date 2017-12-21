package main

import (
	"flag"
	"net/http"
)

var (
	//cache *cacheDb

	httpPort = flag.Int("http-port", 4500, "Http server port")
	httpAddr = flag.String("http-addr", "127.0.0.1", "Http server listen address")
	useHttp  = flag.Bool("use-http", true, "Use http server")
)

func init() {
	//cache = newCacheDb()
}

func main() {
	if *useHttp {
		httpServer := newHttpServer(*httpAddr, *httpPort)
		httpServer.registerHandler(func(writer http.ResponseWriter, request *http.Request) {

		})
	}
}
