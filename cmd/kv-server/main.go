package main

import (
	"flag"
	"log"
	"sync"

	"github.com/2tvenom/kv/kv"
	"github.com/2tvenom/kv/server"
)

var (
	httpPort    = flag.Int("http-port", 4500, "Http server port")
	httpAddr    = flag.String("http-addr", "127.0.0.1", "Http server listen address")
	useHttp     = flag.Bool("use-http", true, "Use http server")
	useTcp      = flag.Bool("use-tcp", true, "Use tcp server")
	useTcpNcat  = flag.Bool("use-tcp-ncat", true, "Use tcp server for ncat client")
	secure      = flag.Bool("secure", false, "Enable TLS auth")
	tcpPortNcat = flag.Int("tcp-port-ncat", 4501, "TCP server port for nncat")
	tcpPort     = flag.Int("tcp-port", 4502, "TCP server port")
	tcpAddr     = flag.String("tcp-addr", "127.0.0.1", "TCP server listen address")
)

func main() {
	cache := kv.NewCacheDb()

	w := sync.WaitGroup{}
	if *useHttp {
		httpServer := server.NewHttpServer(cache, *httpAddr, *httpPort)
		w.Add(1)
		go func() {
			var err error
			if *secure {
				err = httpServer.ListenSecure()
			} else {
				err = httpServer.Listen()
			}
			if err != nil {
				log.Fatalf("Http server error: %s", err.Error())
			}
			w.Done()
		}()
	}

	if *useTcpNcat {
		w.Add(1)

		tcpServer := server.NewTcpServer(cache, *tcpAddr, *tcpPortNcat)
		tcpServer.IsHuman(true)

		go func() {
			err := tcpServer.Listen()

			if err != nil {
				log.Fatalf("TCP server error: %s", err.Error())
			}
			w.Done()
		}()
	}

	if *useTcp {
		w.Add(1)

		tcpServer := server.NewTcpServer(cache, *tcpAddr, *tcpPort)

		go func() {
			var err error
			if *secure {
				err = tcpServer.ListenSecure()
			} else {
				err = tcpServer.Listen()
			}

			if err != nil {
				log.Fatalf("TCP server error: %s", err.Error())
			}
			w.Done()
		}()
	}

	w.Wait()
}
