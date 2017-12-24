package kv

import (
	"testing"
	"time"
	"unsafe"
)

func TestEntryMapping(t *testing.T) {
	elem := &entry{72, uint64(time.Now().Unix()), 3}
	t.Logf("Set elem: %+v; len: %d\n", elem, unsafe.Sizeof(elem))
	data := *(*[headerLen]byte)(unsafe.Pointer(elem))
	t.Logf("Data: %+v", data)

	newElem := (*entry)(unsafe.Pointer(&data[0]))

	t.Logf("get elem: %+v\n", newElem)
}

func TestMappingInt(t *testing.T) {
	var i uint16 = 65535

	data := *(*[2]byte)(unsafe.Pointer(&i))
	t.Logf("Data: %+v", data)

	newElem := *(*uint16)(unsafe.Pointer(&data[0]))

	t.Logf("New: %+v", newElem)
}

func TestTTL(t *testing.T) {
	now := time.Now().Unix()

	t.Logf("Now: %d, TTL: %d", now, getTTL(5))

}

func TestBlockByKey(t *testing.T) {
	t.Logf("Block id: %d", blockByKey("Hello"))
	t.Logf("Block id: %d", blockByKey("test"))
	t.Logf("Block id: %d", blockByKey("foo"))
	t.Logf("Block id: %d", blockByKey("bar"))
	t.Logf("Block id: %d", blockByKey("bar"))
	t.Logf("Block id: %d", blockByKey("bazsdfdsjfiosdjfio"))
}
