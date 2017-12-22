package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

var (
	cache *simpleCacheDb

	httpPort = flag.Int("http-port", 4500, "Http server port")
	httpAddr = flag.String("http-addr", "127.0.0.1", "Http server listen address")
	useHttp  = flag.Bool("use-http", true, "Use http server")
	useTcp   = flag.Bool("use-tcp", true, "Use tcp server")
	secure   = flag.Bool("secure", false, "Enable TLS auth")
	tcpPort  = flag.Int("tcp-port", 4501, "TCP server port")
	tcpAddr  = flag.String("tcp-addr", "127.0.0.1", "TCP server listen address")

)

type (
	output struct {
		Error string      `json:"error,omitempty"`
		Data  interface{} `json:"data,omitempty"`
	}
)

func init() {
	cache = newSimpleCacheDb()
}

func main() {
	w := sync.WaitGroup{}
	if *useHttp {
		httpServer := newHttpServer(*httpAddr, *httpPort)
		httpServer.registerHandler(func(writer http.ResponseWriter, request *http.Request) {
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

			out, err := Exe(cache, parser)
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
		})

		w.Add(1)
		go func() {
			var err error
			if *secure {
				err = httpServer.listenSecure()
			} else {
				err = httpServer.listen()
			}
			if err != nil {
				log.Fatalf("Http server error: %s", err.Error())
			}
			w.Done()
		}()
	}

	if *useTcp {
		w.Add(1)

		tcpServer := newTcpServer(*tcpAddr, *tcpPort)
		tcpServer.registerHandler(func(conn net.Conn) {
			defer conn.Close()

			parser := &baseCommandParser{}
			_, err := io.Copy(parser, conn)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
				return
			}
			log.Printf("CMD: %+v", parser)

			out, err := Exe(cache, parser)
			if err != nil {
				if err == notFoundErr {
					conn.Write([]byte("not found"))
					return
				}
				conn.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
				return
			}

			switch v := out.(type) {
			case string:
				conn.Write([]byte(v))
			case []string:
				for _, e := range v {
					conn.Write([]byte(e + "\n"))
				}
			case map[string]string:
				for k, e := range v {
					conn.Write([]byte(k + "\n"))
					conn.Write([]byte(e + "\n"))
				}
			}
		})

		go func() {
			var err error
			if *secure {
				err = tcpServer.listenSecure()
			} else {
				err = tcpServer.listen()
			}

			if err != nil {
				log.Fatalf("TCP server error: %s", err.Error())
			}
			w.Done()
		}()
	}

	w.Wait()
}
