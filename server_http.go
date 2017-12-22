package main

import (
	"fmt"
	"net/http"
)

type (
	httpServer struct {
		mux    *http.ServeMux
		server *http.Server
	}
)

func newHttpServer(addr string, port int) *httpServer {
	mux := http.NewServeMux()
	return &httpServer{
		mux:    mux,
		server: &http.Server{Addr: fmt.Sprintf("%s:%d", addr, port), Handler: mux},
	}
}

func (s *httpServer) registerHandler(f func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc("/", f)
}

func (s *httpServer) listen() error {
	return s.server.ListenAndServe()
}
