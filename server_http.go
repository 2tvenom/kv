package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type (
	httpServer struct {
		cache  *simpleCacheDb
		server *http.Server
	}

	output struct {
		Error string      `json:"error,omitempty"`
		Data  interface{} `json:"data,omitempty"`
	}
)

func newHttpServer(cache *simpleCacheDb, addr string, port int) *httpServer {
	s := &httpServer{
		cache: cache,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handler)

	s.server = &http.Server{Addr: fmt.Sprintf("%s:%d", addr, port), Handler: mux}
	return s
}

func (s *httpServer) handler(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	e := json.NewEncoder(writer)

	parser := &baseCommandParser{}
	_, err := io.Copy(parser, request.Body)

	log.Printf("CMD %+v\n", parser)

	if err != nil {
		e.Encode(&output{Error: err.Error()})
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	out, err := Exe(s.cache, parser)
	if err != nil {
		if err == notFoundErr {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		e.Encode(&output{Error: err.Error()})
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	e.Encode(&output{Data: out})
	return
}

func (s *httpServer) listenSecure() error {
	cert, key, cfg := getTLSConfig()
	s.server.TLSConfig = cfg

	return s.server.ListenAndServeTLS(cert, key)
}

func (s *httpServer) listen() error {
	return s.server.ListenAndServe()
}
