package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"encoding/binary"
)

type (
	Client struct {
		conn net.Conn
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

func NewClient(addr string, port int) (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Do(cmd string) (interface{}, error) {
	data := uint32ToBytesClientConvert(uint32(len(cmd)))
	_, err := c.conn.Write(append([]byte{header}, data...))
	if err != nil {
		return nil, err
	}

	buff := strings.NewReader(cmd)
	_, err = io.Copy(c.conn, buff)

	if err != nil {
		return nil, err
	}

	header := make([]byte, 1)
	_, err = c.conn.Read(header)
	if err != nil {
		return nil, err
	}

	//log.Printf("Get % x:", header)
	switch header[0] {
	case ok:
		dataType := make([]byte, 1)
		_, err = c.conn.Read(dataType)
		if err != nil {
			return nil, err
		}

		switch dataType[0] {
		case typeNone:
			return true, nil
		case typeString:
			buff, err := readData(c.conn)
			if err != nil {
				return nil, err
			}
			return string(buff), nil
		case typeList:
			cnt, err := readUInt(c.conn)
			if err != nil {
				return err, nil
			}

			out := make([]string, cnt)
			for i := 0; i < int(cnt); i++ {
				buff, err := readData(c.conn)
				if err != nil {
					return nil, err
				}
				out[i] = string(buff)
			}

			return out, nil
		case typeDict:
			cnt, err := readUInt(c.conn)
			if err != nil {
				return err, nil
			}

			out := map[string]string{}

			for i := 0; i < int(cnt); i++ {
				key, err := readData(c.conn)
				if err != nil {
					return nil, err
				}
				value, err := readData(c.conn)
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
		errBuff, err := readData(c.conn)
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
