package main

import (
	"flag"
	"net/http"
	"io"
	"encoding/json"
	"log"
	"net"
	"fmt"
	"sync"
)

var (
	cache *simpleCacheDb

	httpPort = flag.Int("http-port", 4500, "Http server port")
	httpAddr = flag.String("http-addr", "127.0.0.1", "Http server listen address")
	useHttp  = flag.Bool("use-http", true, "Use http server")
	useTcp   = flag.Bool("use-tcp", true, "Use tcp server")
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

			log.Printf("OUT %+v\n", out)

			e.Encode(&output{Data: out})
			return
		})

		w.Add(1)
		go func() {
			err := httpServer.listen()
			if err != nil {
				log.Fatalf("Http server error: %s", err.Error())
			}
			w.Done()
		}()
	}

	if *useTcp {
		w.Add(1)
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *tcpAddr, *tcpPort))
		if err != nil {
			log.Fatalf("Error listening: %s", err.Error())
		}
		// Close the listener when the application closes.
		defer l.Close()
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				log.Fatalf("Error accepting: %s", err.Error())
			}

			go func() {
				defer conn.Close()
				log.Println("GOT CONN")

				parser := &baseCommandParser{}
				_, err := io.Copy(parser, conn)
				if err != nil {
					log.Println("GOT ERR", err.Error())
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

				fmt.Printf("%+v\n", out)

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

			}()
		}
	}

	w.Wait()
}
