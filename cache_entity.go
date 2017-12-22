package main

import (
	"errors"
	"hash/fnv"
	"time"
)

type (
	entry struct {
		length  uint64
		ttl     uint64
		keyType uint8
	}
)

const (
	blocks = 256

	headerLen        = 17
	maxListElemennts = (1 << 16) - 1

	keyString = 1
	keyList   = 2
	keyDict   = 3
)

var (
	notFoundErr             = errors.New("Not found")
	incorrectSelectKeyType  = errors.New("Incorrect select key type")
	incorrectDictElementErr = errors.New("Incorrect dictionary element")
)

func blockByKey(key string) uint8 {
	hash := fnv.New64()
	hash.Write([]byte(key))
	sum := hash.Sum64()
	return uint8(sum & 255)
}

func getTTL(ttl int64) uint64 {
	if ttl == 0 {
		return 0
	}

	return uint64(time.Now().Unix() + ttl)
}
