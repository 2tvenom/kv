package kv

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"
	"unsafe"
)

type (
	CacheDb struct {
		blocks [blocks]map[string][]byte
		locks  [blocks]sync.RWMutex
	}
)

var (
	_ = fmt.Printf
)

func NewCacheDb() *CacheDb {
	c := &CacheDb{}
	for i := 0; i < blocks; i++ {
		c.blocks[i] = map[string][]byte{}
	}
	return c
}

func (c *CacheDb) Keys() []string {
	out := []string{}
	for i, block := range c.blocks {
		c.locks[i].Lock()
		for key := range block {
			out = append(out, key)
		}
		c.locks[i].Unlock()
	}
	return out
}

func (c *CacheDb) Remove(key string) {
	id := blockByKey(key)
	c.locks[id].Lock()
	delete(c.blocks[id], key)
	c.locks[id].Unlock()
}

func (c *CacheDb) get(key string, keyType uint8) ([]byte, error) {
	id := blockByKey(key)
	c.locks[id].RLock()
	if data, ok := c.blocks[id][key]; ok {
		header := make([]byte, headerLen)
		copy(header, data[:headerLen])
		entry := (*entry)(unsafe.Pointer(&header[0]))
		if entry.keyType != keyType {
			return nil, incorrectSelectKeyType
		}
		now := time.Now().Unix()

		if entry.ttl > 0 && entry.ttl <= uint64(now) {
			c.locks[id].RUnlock()
			c.locks[id].Lock()
			delete(c.blocks[id], key)
			c.locks[id].Unlock()
			return nil, notFoundErr
		}
		out := make([]byte, entry.length)
		copy(out, data[headerLen:])
		c.locks[id].RUnlock()
		return out, nil
	} else {
		c.locks[id].RUnlock()
		return nil, notFoundErr
	}
}

func (c *CacheDb) set(key string, keyType uint8, ttl int64, value []byte) error {
	elem := &entry{uint64(len(value)), getTTL(ttl), keyType}
	data := *(*[headerLen]byte)(unsafe.Pointer(elem))

	nVal := make([]byte, len(value))
	copy(nVal, value)

	id := blockByKey(key)
	c.locks[id].Lock()
	c.blocks[id][key] = append(data[:], nVal...)
	c.locks[id].Unlock()
	return nil
}

func (c *CacheDb) setList(key string, keyType uint8, ttl int64, values [][]byte) error {
	if len(values) > maxListElemennts {
		return tooMatchListElementsErr
	}
	off := (len(values) * 2) + 2
	lenBuff := off
	for _, val := range values {
		lenBuff += len(val)
	}

	if keyType == keyDict {
		lenBuff += len(values) * 2
	}

	buff := make([]byte, lenBuff)
	//write counnt elements in record header
	data := sliceUnsafeConvert(uint16(len(values)))
	copy(buff[:2], data[:])

	for i, val := range values {
		elemLen := len(val)
		if keyType == keyDict {
			elemLen += 2
		}
		//write elem length in record header
		data := sliceUnsafeConvert(uint16(elemLen))
		copy(buff[i*2+2:(i*2)+4], data[:])
		if keyType == keyDict {
			//additional separator index in record start
			sepIndex := uint16(bytes.Index(val, dictionarySeparator))
			sepData := sliceUnsafeConvert(uint16(sepIndex))
			copy(buff[off:off+2], sepData[:])
			copy(buff[off+2:off+elemLen], val)
		} else {
			copy(buff[off:off+elemLen], val)
		}
		off += elemLen
	}

	return c.set(key, keyType, ttl, buff)
}

func (c *CacheDb) SetList(key string, ttl int64, values [][]byte) error {
	return c.setList(key, keyList, ttl, values)
}

func (c *CacheDb) Get(key string) ([]byte, error) {
	return c.get(key, keyString)
}

func (c *CacheDb) Set(key string, ttl int64, value []byte) error {
	return c.set(key, keyString, ttl, value)
}

func (c *CacheDb) getList(key string, keyType uint8) ([][]byte, error) {
	data, err := c.get(key, keyType)
	if err != nil {
		return nil, err
	}

	elemCount := uint16UnsafeConvert(data)
	out := make([][]byte, elemCount)

	off := (elemCount * 2) + 2
	var i uint16
	for i = 0; i < elemCount; i++ {
		elemLen := uint16UnsafeConvert(data[i*2+2: i*2+4])
		out[i] = make([]byte, elemLen)
		copy(out[i], data[off:off+elemLen])
		off += elemLen
	}

	return out, nil
}

func (c *CacheDb) GetList(key string) ([][]byte, error) {
	return c.getList(key, keyList)
}

func (c *CacheDb) GetListElement(key string, position uint16) ([]byte, error) {
	data, err := c.get(key, keyList)

	if err != nil {
		return nil, err
	}

	off, elemLen, err := getElemByPosition(data, position)
	if err != nil {
		return nil, err
	}
	return data[off: off+elemLen], nil
}

func (c *CacheDb) SetDict(key string, ttl int64, values dictionary) error {
	for _, val := range values {
		if bytes.Index(val, dictionarySeparator) == -1 {
			return incorrectDictElementErr
		}
	}
	sort.Sort(values)

	return c.setList(key, keyDict, ttl, values)
}

func (c *CacheDb) GetDict(key string) ([][]byte, error) {
	return c.getList(key, keyDict)
}

func (c *CacheDb) GetDictElement(key string, dictKey []byte) ([]byte, error) {
	data, err := c.get(key, keyDict)

	if err != nil {
		return nil, err
	}

	elemCount := int(uint16UnsafeConvert(data))

	//custom binary search
	i := sort.Search(elemCount, func(position int) bool {
		off, elemLen, _ := getElemByPosition(data, uint16(position))

		separatorPosition := uint64(uint16UnsafeConvert(data[off: off+elemLen]))
		return bytes.Compare(data[off+2:off+2+separatorPosition], dictKey) >= 0
	})

	if i < elemCount {
		off, elemLen, _ := getElemByPosition(data, uint16(i))
		separatorPosition := uint64(uint16UnsafeConvert(data[off: off+elemLen]))
		if bytes.Equal(data[off+2:off+2+separatorPosition], dictKey) {
			return data[off+2+1+separatorPosition: off+elemLen], nil
		} else {
			return nil, notFoundErr
		}
	} else {
		return nil, notFoundErr
	}
}

func uint16UnsafeConvert(data []byte) uint16 {
	elemCountData := make([]byte, 2)
	copy(elemCountData, data[0:2])
	return *(*uint16)(unsafe.Pointer(&elemCountData[0]))
}

func sliceUnsafeConvert(val uint16) []byte {
	return (*(*[2]byte)(unsafe.Pointer(&val)))[:]
}

//get element by position in byte slice
func getElemByPosition(data []byte, position uint16) (uint64, uint64, error) {
	elemCount := uint16UnsafeConvert(data)
	if position >= elemCount {
		return 0, 0, notFoundErr
	}

	off := uint64(elemCount)*2 + 2
	var i uint16
	for i = 0; i < position; i++ {
		off += uint64(uint16UnsafeConvert(data[i*2+2: i*2+4]))
	}

	elemLen := uint16UnsafeConvert(data[i*2+2: i*2+4])
	return off, uint64(elemLen), nil
}
