package main

import (
	"hash/fnv"
	"time"
	"errors"
)

type (
	entry struct {
		length uint64
		ttl    uint64
	}
)

var (
	notFoundErr = errors.New("Not found")
)

func blockByKey(key string) uint8 {
	hash := fnv.New64()
	hash.Write([]byte(key))
	sum := hash.Sum64()
	return uint8(sum & 255)
}

func ttl(ttl int64) int64 {
	if ttl == 0 {
		return 0
	}

	return time.Now().Unix() + ttl
}
