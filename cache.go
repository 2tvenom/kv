package main

import (
	"sync"

	"unsafe"
	"time"
)

const (
	segments = 256
)

type (
	simpleCacheDb struct {
		blocks [segments]map[string][]byte
		locks  [segments]sync.RWMutex
	}
)

func newSimpleCacheDb() *simpleCacheDb {
	c := &simpleCacheDb{}
	for i := 0; i < segments; i++ {
		c.blocks[i] = map[string][]byte{}
	}
	return c
}

func (c *simpleCacheDb) Get(key string) ([]byte, error) {
	id := blockByKey(key)
	c.locks[id].RLock()
	if data, ok := c.blocks[id][key]; ok {
		e := data[:16]
		entry := (*entry)(unsafe.Pointer(&e[0]))
		if entry.ttl > 0 && entry.ttl <= uint64(time.Now().Unix()) {
			c.locks[id].RUnlock()
			c.locks[id].Lock()
			delete(c.blocks[id], key)
			c.locks[id].Unlock()
			return nil, notFoundErr
		}
		c.locks[id].RUnlock()
		return data[16:], nil
	} else {
		c.locks[id].RUnlock()
		return nil, notFoundErr
	}
}

func (c *simpleCacheDb) Set(key string, ttl uint64, value []byte) bool {
	elem := &entry{uint64(len(value)), ttl}
	data := *(*[16]byte)(unsafe.Pointer(elem))

	id := blockByKey(key)
	c.locks[id].Lock()
	c.blocks[id][key] = append(data[:], value...)
	c.locks[id].Unlock()
	return true
}
