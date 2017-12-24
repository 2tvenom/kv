package client

import (
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type (
	Client struct {
		sync.Mutex

		conns *list.List
		addr  string
		port  int

		maxIdleConns int
	}
)

const (
	header   = 0x11
	ok       = 0x22
	notFound = 0x44
	errHead  = 0x99

	typeNone   = 0x50
	typeString = 0x51
	typeList   = 0x52
	typeDict   = 0x53
)

var (
	NotFoundErr = errors.New("Not found")
)

type PoolConn struct {
	net.Conn
	c *Client
}

func (c *PoolConn) Close() {
	c.c.put(c.Conn)
}

func (c *PoolConn) Finalize() {
	c.Conn.Close()
}

func NewClient(addr string, port int) *Client {
	return &Client{
		addr:         addr,
		port:         port,
		maxIdleConns: 16,
		conns:        list.New(),
	}
}

func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()

	for c.conns.Len() > 0 {
		e := c.conns.Front()
		co := e.Value.(net.Conn)
		c.conns.Remove(e)
		co.Close()
	}
}

func (c *Client) Get() (*PoolConn, error) {
	co, err := c.get()
	if err != nil {
		return nil, err
	}

	return &PoolConn{co, c}, err
}

func (c *Client) newConn() (co net.Conn, err error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.addr, c.port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
func (c *Client) get() (co net.Conn, err error) {
	c.Lock()
	if c.conns.Len() == 0 {
		c.Unlock()
		co, err = c.newConn()
	} else {
		e := c.conns.Front()
		co = e.Value.(net.Conn)
		c.conns.Remove(e)

		c.Unlock()
	}

	return
}

func (c *Client) put(conn net.Conn) {
	c.Lock()
	defer c.Unlock()

	for c.conns.Len() >= c.maxIdleConns {
		// remove back
		e := c.conns.Back()
		co := e.Value.(net.Conn)
		c.conns.Remove(e)
		co.Close()
	}

	c.conns.PushFront(conn)
}

func (c *Client) Do(cmd string) (interface{}, error) {
	conn, err := c.get()
	if err != nil {
		return nil, err
	}

	defer c.put(conn)

	data := uint32ToBytesClientConvert(uint32(len(cmd)))
	_, err = conn.Write(append([]byte{header}, data...))
	if err != nil {
		return nil, err
	}

	buff := strings.NewReader(cmd)
	_, err = io.Copy(conn, buff)

	if err != nil {
		return nil, err
	}

	header := make([]byte, 1)
	_, err = conn.Read(header)
	if err != nil {
		return nil, err
	}

	//log.Printf("Get % x:", header)
	switch header[0] {
	case ok:
		dataType := make([]byte, 1)
		_, err = conn.Read(dataType)
		if err != nil {
			return nil, err
		}

		switch dataType[0] {
		case typeNone:
			return true, nil
		case typeString:
			buff, err := readData(conn)
			if err != nil {
				return nil, err
			}
			return string(buff), nil
		case typeList:
			cnt, err := readUInt(conn)
			if err != nil {
				return err, nil
			}

			out := make([]string, cnt)
			for i := 0; i < int(cnt); i++ {
				buff, err := readData(conn)
				if err != nil {
					return nil, err
				}
				out[i] = string(buff)
			}

			return out, nil
		case typeDict:
			cnt, err := readUInt(conn)
			if err != nil {
				return err, nil
			}

			out := map[string]string{}

			for i := 0; i < int(cnt); i++ {
				key, err := readData(conn)
				if err != nil {
					return nil, err
				}
				value, err := readData(conn)
				if err != nil {
					return nil, err
				}
				out[string(key)] = string(value)
			}

			return out, nil
		}

		return nil, nil
	case notFound:
		return nil, NotFoundErr
	case errHead:
		errBuff, err := readData(conn)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(errBuff))
	default:
		return nil, nil
	}
}

func bytesToUint32ClientConvert(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[0:4])
}

func uint32ToBytesClientConvert(v uint32) []byte {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint32(out, v)

	return out
}

func readUInt(conn net.Conn) (uint32, error) {
	lenHeader := make([]byte, 4)
	_, err := conn.Read(lenHeader)
	if err != nil {
		return 0, err
	}
	return bytesToUint32ClientConvert(lenHeader), nil
}

func readData(conn net.Conn) ([]byte, error) {
	l, err := readUInt(conn)
	if err != nil {
		return nil, err
	}

	buff := make([]byte, int(l))
	_, err = conn.Read(buff)
	if err != nil {
		return nil, err
	}

	return buff, nil
}
