package main

import (
	"fmt"
	"net"
	"crypto/tls"
)

type (
	tcpServer struct {
		handler func(conn net.Conn)
		addr    string
		port    int
	}
)

func newTcpServer(addr string, port int) *tcpServer {
	return &tcpServer{
		addr: addr,
		port: port,
	}
}

func (s *tcpServer) registerHandler(f func(conn net.Conn)) {
	s.handler = f
}

func (s *tcpServer) listenSecure() error {
	_, _, cfg := getTLSConfig()
	l, err := tls.Listen("tcp", fmt.Sprintf("%s:%d", s.addr, s.port), cfg)
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			conn.Close()
			continue
		}
		go s.handler(conn)
	}
	return err
}

func (s *tcpServer) listen() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.addr, s.port))

	// Close the listener when the application closes.
	defer l.Close()
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			conn.Close()
			continue
		}
		go s.handler(conn)
	}
	return err
}
