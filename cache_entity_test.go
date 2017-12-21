package main

import (
	"testing"
	"unsafe"

)

func TestEntryMapping(t *testing.T) {
	//val := []byte("Hello")
	elem := &entry{12, 1, 10}

	data := *(*[headerLen]byte)(unsafe.Pointer(elem))
	t.Logf("Data: %+v", data)

	newElem := (*entry)(unsafe.Pointer(&data[0]))

	t.Logf("New: %+v", newElem)
}

func TestMappingInt(t *testing.T) {

	var i uint16 = 65535

	data := *(*[2]byte)(unsafe.Pointer(&i))
	t.Logf("Data: %+v", data)

	newElem := *(*uint16)(unsafe.Pointer(&data[0]))

	t.Logf("New: %+v", newElem)
}

func TestBlockByKey(t *testing.T) {
	t.Logf("Block id: %d", blockByKey("Hello"))
	t.Logf("Block id: %d", blockByKey("test"))
	t.Logf("Block id: %d", blockByKey("foo"))
	t.Logf("Block id: %d", blockByKey("bar"))
	t.Logf("Block id: %d", blockByKey("bar"))
	t.Logf("Block id: %d", blockByKey("bazsdfdsjfiosdjfio"))
}