package main

import (
	"bytes"
	"strconv"
	"errors"
)

func Exe(cache *simpleCacheDb, parser *baseCommandParser) (interface{}, error) {
	parser.value = bytes.TrimSpace(parser.value)
	if !parser.headerParsed {
		return nil, errors.New("Incorrect command")
	}
	//log.Printf("CMD %+v", parser)
	switch parser.cmd {
	case cmdGetLex:
		data, err := cache.Get(parser.key)
		if err != nil {
			return nil, err
		}
		return string(data), nil
	case cmdGetListLex:
		data, err := cache.GetList(parser.key)
		if err != nil {
			return nil, err
		}

		out := make([]string, len(data))
		for i, elem := range data {
			out[i] = string(elem)
		}
		return out, nil
	case cmdGetListElemLex:
		i, err := strconv.Atoi(string(parser.value))
		if err != nil {
			return nil, err
		}
		data, err := cache.GetListElement(parser.key, uint16(i))
		if err != nil {
			return nil, err
		}
		return string(data), nil
	case cmdGetDictLex:
		data, err := cache.GetDict(parser.key)
		if err != nil {
			return nil, err
		}

		out := map[string]string{}
		for _, elem := range data {
			indexPos := uint16UnsafeConvert(elem)
			out[string(elem[2:indexPos+2])] = string(elem[indexPos+1+2:])
		}
		return out, nil
	case cmdGetDictElemLex:
		data, err := cache.GetDictElement(parser.key, parser.value)
		if err != nil {
			return nil, err
		}
		return string(data), nil
	case cmdSetLex:
		return nil, cache.Set(parser.key, parser.ttl, parser.value)
	case cmdSetListLex:
		return nil, cache.SetList(parser.key, parser.ttl, bytes.Split(parser.value, []byte(" ")))
	case cmdSetDictLex:
		return nil, cache.SetDict(parser.key, parser.ttl, bytes.Split(parser.value, []byte(" ")))
	case cmdKeysLex:
		return cache.Keys(), nil
	case cmdRemoveLex:
		cache.Remove(parser.key)
		return nil, nil
	default:
		return nil, notFoundErr
	}
}
