package main

import (
	"flag"
	"log"
	"sync"
)

var (
	httpPort = flag.Int("http-port", 4500, "Http server port")
	httpAddr = flag.String("http-addr", "127.0.0.1", "Http server listen address")
	useHttp  = flag.Bool("use-http", true, "Use http server")
	useTcp   = flag.Bool("use-tcp", true, "Use tcp server")
	secure   = flag.Bool("secure", false, "Enable TLS auth")
	tcpPort  = flag.Int("tcp-port", 4501, "TCP server port")
	tcpAddr  = flag.String("tcp-addr", "127.0.0.1", "TCP server listen address")
)

func main() {
	cache := newSimpleCacheDb()

	w := sync.WaitGroup{}
	if *useHttp {
		httpServer := newHttpServer(cache, *httpAddr, *httpPort)
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

		tcpServer := newTcpServer(cache, *tcpAddr, *tcpPort)

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
