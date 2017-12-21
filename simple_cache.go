package main

import (
	"unsafe"
	"time"
)

type (
	simpleCacheDb struct {
		blocks [blocks]map[string][]byte
		lockers
	}
)

func newSimpleCacheDb() *simpleCacheDb {
	c := &simpleCacheDb{}
	for i := 0; i < blocks; i++ {
		c.blocks[i] = map[string][]byte{}
	}
	return c
}

func (c *simpleCacheDb) Get(key string) ([]byte, error) {
	return c.get(key, keyString)
}

func (c *simpleCacheDb) get(key string, keyType uint8) ([]byte, error) {
	id := blockByKey(key)
	c.RLock(id)
	if data, ok := c.blocks[id][key]; ok {
		e := data[:headerLen]
		entry := (*entry)(unsafe.Pointer(&e[0]))
		if entry.ttl > 0 && entry.ttl <= uint64(time.Now().Unix()) {
			c.RUnlock(id)
			c.Lock(id)
			delete(c.blocks[id], key)
			c.Unlock(id)
			return nil, notFoundErr
		}
		out := make([]byte, entry.length)
		copy(out, data[headerLen:])
		c.RUnlock(id)
		return out, nil
	} else {
		c.RUnlock(id)
		return nil, notFoundErr
	}
}

func (c *simpleCacheDb) Set(key string, ttl uint64, value []byte) bool {
	return c.set(key, keyString, ttl, value)
}

func (c *simpleCacheDb) set(key string, keyType uint8, ttl uint64, value []byte) bool {
	elem := &entry{uint64(len(value)), keyType, ttl}
	data := *(*[headerLen]byte)(unsafe.Pointer(elem))

	nVal := make([]byte, len(value))
	copy(nVal, value)

	id := blockByKey(key)
	c.Lock(id)
	c.blocks[id][key] = append(data[:], nVal...)
	c.Unlock(id)
	return true
}

func (c *simpleCacheDb) SetList(key string, ttl uint64, values [][]byte) bool {
	off := (len(values) * 2) + 2
	lenBuff := off
	for _, val := range values {
		lenBuff += len(val)
	}
	buff := make([]byte, lenBuff)
	countElem := len(values)
	data := *(*[2]byte)(unsafe.Pointer(&countElem))
	copy(buff[:2], data[:])

	for i, val := range values {
		elemLen := len(val)
		data := *(*[2]byte)(unsafe.Pointer(&elemLen))
		copy(buff[i*2+2:(i*2)+4], data[:])
		copy(buff[off:off+elemLen], val)
		off += elemLen
	}

	return c.set(key, keyList, ttl, buff)
}

func (c *simpleCacheDb) GetList(key string) ([][]byte, error) {
	data, err := c.get(key, keyList)

	if err != nil {
		return nil, err
	}

	elemCount := *(*uint16)(unsafe.Pointer(&data[0]))
	out := make([][]byte, elemCount)

	off := (elemCount * 2) + 2
	var i uint16
	for i = 0; i < elemCount; i++ {
		elemLen := *(*uint16)(unsafe.Pointer(&data[i*2+2:i*2+4][0]))
		out[i] = make([]byte, elemLen)
		copy(out[i], data[off:off+elemLen])
		println(string(out[i]))
		off += elemLen
	}

	return out, nil
}
