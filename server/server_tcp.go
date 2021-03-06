package server

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"crypto/rand"
	"time"

	"github.com/2tvenom/kv/kv"
)

type (
	tcpServer struct {
		cache           *kv.CacheDb
		addr            string
		port            int
		isHumanListener bool
	}
)

const (
	clientHeader   = 0x11
	okHeader       = 0x22
	notFoundHeader = 0x44
	errHeader      = 0x99

	dataTypeNone   = 0x50
	dataTypeString = 0x51
	dataTypeList   = 0x52
	dataTypeDict   = 0x53
)

func NewTcpServer(cache *kv.CacheDb, addr string, port int) *tcpServer {
	return &tcpServer{
		addr:  addr,
		port:  port,
		cache: cache,
	}
}

func (s *tcpServer) IsHuman(b bool) {
	s.isHumanListener = b
}

func (s *tcpServer) humanHandler(conn net.Conn) {
	defer conn.Close()

	parser := &baseCommandParser{}
	_, err := io.Copy(parser, conn)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
		return
	}
	//log.Printf("CMD: %+v", parser)

	out, err := Exe(s.cache, parser)
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
}

func (s *tcpServer) clientHandler(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Minute))
	conn.SetWriteDeadline(time.Now().Add(time.Minute))

	for {
		header := make([]byte, 5)
		n, err := conn.Read(header)
		if err != nil {
			return
		}

		if n != 5 {
			_, err := conn.Write(errPack(errors.New(fmt.Sprintf("Incorrect length read header, expected 5 got %d", n))))
			if err != nil {
				return
			}
			continue
		}

		if header[0] != clientHeader {
			return
		}

		requestLen := bytesToUint32Convert(header[1:])
		parser := &baseCommandParser{}
		n1, err := io.CopyN(parser, conn, int64(requestLen))
		if n1 != int64(requestLen) {
			_, err := conn.Write(errPack(errors.New("Incorrect length read body")))
			if err != nil {
				return
			}
			continue
		}

		//log.Printf("CMD LEN: %d %d [% x]", requestLen, n, header[1:5])
		if err != nil {
			_, err := conn.Write(errPack(err))
			if err != nil {
				return
			}
			continue
		}

		//log.Printf("CMD: %+v %+v", parser, err)

		out, err := Exe(s.cache, parser)
		if err != nil {
			if err == notFoundErr {
				_, err := conn.Write([]byte{notFoundHeader})
				if err != nil {
					return
				}
				continue
			}
			_, err = conn.Write(errPack(err))
			if err != nil {
				return
			}
			continue
		}

		switch data := out.(type) {
		case string:
			buff := []byte{okHeader, dataTypeString}
			lenPack := uint32ToBytesConvert(uint32(len(data)))
			buff = append(buff, lenPack...)
			buff = append(buff, []byte(data)...)

			_, err := conn.Write(buff)
			if err != nil {
				return
			}
		case []string:
			lenPack := uint32ToBytesConvert(uint32(len(data)))
			_, err := conn.Write(append([]byte{okHeader, dataTypeList}, lenPack...))
			if err != nil {
				return
			}

			buff := []byte{}
			for _, e := range data {
				lenPack := uint32ToBytesConvert(uint32(len(e)))
				buff = append(buff, lenPack...)
				buff = append(buff, []byte(e)...)
			}

			_, err = conn.Write(buff)
			if err != nil {
				return
			}

		case map[string]string:
			lenPack := uint32ToBytesConvert(uint32(len(data)))
			_, err := conn.Write(append([]byte{okHeader, dataTypeDict}, lenPack...))
			if err != nil {
				return
			}

			buff := []byte{}
			for k, v := range data {
				lenPack := uint32ToBytesConvert(uint32(len(k)))
				buff = append(buff, lenPack...)
				buff = append(buff, []byte(k)...)

				lenPack = uint32ToBytesConvert(uint32(len(v)))
				buff = append(buff, lenPack...)
				buff = append(buff, []byte(v)...)
			}

			_, err = conn.Write(buff)
			if err != nil {
				return
			}
		default:
			conn.Write([]byte{okHeader, dataTypeNone})
		}
	}
}

func (s *tcpServer) listenServ(l net.Listener) error {
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			conn.Close()
			continue
		}
		if s.isHumanListener {
			go s.humanHandler(conn)
		} else {
			go s.clientHandler(conn)
		}
	}
}

func (s *tcpServer) ListenSecure(certPath string, keyPath string) error {
	tlsConfig, err := getTLS(certPath)
	if err != nil {
		return nil
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}

	tlsConfig.Certificates = []tls.Certificate{cert}
	tlsConfig.Rand = rand.Reader

	l, err := tls.Listen("tcp", fmt.Sprintf("%s:%d", s.addr, s.port), tlsConfig)
	if err != nil {
		return err
	}

	defer l.Close()
	s.listenServ(l)
	return nil
}

func (s *tcpServer) Listen() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.addr, s.port))
	if err != nil {
		return err
	}
	defer l.Close()
	s.listenServ(l)

	return nil
}

func errPack(err error) []byte {
	errResponse := []byte(fmt.Sprintf("Error: %s", err.Error()))
	errLength := uint32ToBytesConvert(uint32(len(errResponse)))

	out := append([]byte{errHeader}, errLength...)
	return append(out, errResponse...)
}

func bytesToUint32Convert(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[0:4])
}

func uint32ToBytesConvert(v uint32) []byte {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint32(out, v)

	return out
}
