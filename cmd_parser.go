package main

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"strconv"
	"text/scanner"
	//"hash/fnv"
)

const (
	minBuffSize  = 512
	maxKeyLength = 256
	maxTTLLength = 19

	cmdKeys = iota << 1
	cmdRemove
	cmdGet
	cmdGetList
	cmdGetDict
	cmdSet
	cmdSetList
	cmdSetDict
	cmdGetListElem
	cmdGetDictElem

	cmdKeysLex        = "KEYS"
	cmdRemoveLex      = "REMOVE"
	cmdGetLex         = "GET"
	cmdGetListLex     = "GETLIST"
	cmdGetListElemLex = "GETLISTELEM"
	cmdGetDictLex     = "GETDICT"
	cmdGetDictElemLex = "GETDICTELEM"
	cmdSetLex         = "SET"
	cmdSetListLex     = "SETLIST"
	cmdSetDictLex     = "SETDICT"
)

var (
	approvedCommands      map[string]int
	incorrectCommandError = errors.New("Incorrect command name")
	longKeyNameError      = errors.New("Maximum key name length is 256")
	notTTl                = errors.New("")
	zeroTTl               = errors.New("TTL can not be zero")
)

func init() {
	approvedCommands = map[string]int{
		cmdKeysLex:        cmdKeys,
		cmdRemoveLex:      cmdRemove,
		cmdGetLex:         cmdGet,
		cmdGetListLex:     cmdGetList,
		cmdGetDictLex:     cmdGetDict,
		cmdSetLex:         cmdSet,
		cmdSetListLex:     cmdSetList,
		cmdSetDictLex:     cmdSetDict,
		cmdGetListElemLex: cmdGetListElem,
		cmdGetDictElemLex: cmdGetDictElem,
	}
}

type (
	baseCommandParser struct {
		cmd          string
		key          string
		ttl          int64
		value        []byte
		headerParsed bool
	}
)

//COMMAND key [TTL] value
func (r *baseCommandParser) Write(p []byte) (n int, err error) {
	if !r.headerParsed {
		var s scanner.Scanner
		s.Init(bytes.NewBuffer(p))

		//scan command
		tok := s.Scan()
		if tok == scanner.EOF {
			return s.Offset, io.EOF
		}

		//check command
		cmd := s.TokenText()
		var cmdIndex int
		var ok bool
		if cmdIndex, ok = approvedCommands[cmd]; !ok {
			return 0, incorrectCommandError
		}

		r.cmd = cmd
		//return if CMD = KEYS
		if cmd == cmdKeysLex {
			r.headerParsed = true
			return len(p), nil
		}

		//scan key name
		tok = s.Scan()
		if tok == scanner.EOF {
			return s.Offset, io.EOF
		}

		r.key = s.TokenText()
		if len(r.key) > maxKeyLength {
			return 0, longKeyNameError
		}

		//return if GET(s) commands
		if cmdIndex > cmdKeys && cmdIndex < cmdSet {
			r.headerParsed = true
			return len(p), nil
		}

		//scan ttl if exist
		tok = s.Scan()
		if tok == scanner.EOF {
			return s.Offset, io.EOF
		}

		ttlOffset := s.Offset + maxTTLLength + 2

		if len(p) < ttlOffset {
			ttlOffset = len(p)
		}

		ttl, offset, err := parseTTL(p[s.Offset:ttlOffset])
		switch err {
		case nil:
			s.Offset += offset
			r.ttl = ttl
			tok = s.Scan()
			if tok == scanner.EOF {
				return s.Offset, io.EOF
			}
		case notTTl:
		case zeroTTl:
			fallthrough
		default:

			return 0, err
		}

		//just write value
		r.value = p[s.Offset:]
		r.headerParsed = true
	} else {
		r.value = append(r.value, p...)
	}

	return len(p), nil
}

func parseTTL(p []byte) (int64, int, error) {
	var num []byte

	var hasWhitespace bool
loop:
	for _, b := range p {
		switch {
		case b == 32:
			hasWhitespace = true
			break loop
		case b >= 48 && b <= 57:
			num = append(num, b)
		default:
			return 0, 0, notTTl
		}
	}

	if !hasWhitespace || len(num) == 0 {
		return 0, 0, notTTl
	}

	ttl, err := strconv.Atoi(string(num))
	if err != nil {
		return 0, 0, errors.New("Incorrect ttl: " + err.Error())
	}

	if ttl == 0 {
		return 0, 0, zeroTTl
	}

	return int64(ttl), 1 + len(num), nil
}
