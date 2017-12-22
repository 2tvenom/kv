package main

import (
	"testing"
	"time"
)

func TestGetSetSimpleCache(t *testing.T) {
	cache := newSimpleCacheDb()

	val := "baz"
	cache.Set("foo", 0, []byte(val))

	data, err := cache.Get("foo")
	if err != nil {
		t.Fatal("Get key error", err.Error())
	}

	if string(data) != val {
		t.Fatal("Incorrect value", "expected", val, "got", string(data))
	}
}

func TestGetSetTTLSimpleCache(t *testing.T) {
	cache := newSimpleCacheDb()

	val := "baz"
	cache.Set("foo", 2, []byte(val))
	time.Sleep(time.Second * 3)

	data, err := cache.Get("foo")
	if err == nil {
		t.Fatal("Expected error", "got nil", "data", data)
	}

	if err != notFoundErr {
		t.Fatal("Incorrect error", "expected", notFoundErr, "got", err)
	}
}

func TestGetSetListCache(t *testing.T) {
	cache := newSimpleCacheDb()

	val := [][]byte{
		[]byte("baz"),
		[]byte("bar"),
		[]byte("foo"),
		[]byte("foobaz"),
		[]byte("foo_bar"),
		[]byte("hello"),
	}

	cache.SetList("foo", 0, val)

	data, err := cache.GetList("foo")
	if err != nil {
		t.Fatal("Get key error", err.Error())
	}

	for i, elem := range data {
		if string(elem) != string(val[i]) {
			t.Fatal("Incorrect element", "expected", string(val[i]), "got", string(elem))
		}
	}

	positions := []uint16{3, 0, 2, 5}
	for _, pos := range positions {
		//t.Log(pos)
		elem, err := cache.GetListElement("foo", pos)
		if err != nil {
			t.Fatal("Get key error", err.Error(), pos)
		}

		if string(elem) != string(val[pos]) {
			t.Fatal("Incorrect element", "expected", string(val[pos]), "got", string(elem))
		}
	}

	_, err = cache.GetListElement("foo", 6)
	if err != notFoundErr {
		t.Fatal("Expected error", notFoundErr.Error(), "got", err)
	}
}

func TestGetSetDictCache(t *testing.T) {
	cache := newSimpleCacheDb()

	val := [][]byte{
		[]byte("foo:baz"),
		[]byte("fooo:baz"),
		[]byte("baz:foobaz"),
		[]byte("zbaz:world"),
		[]byte("bar:BAR"),
		[]byte("c:hello"),
		[]byte("d:hello"),
	}

	cache.SetDict("foo", 0, val)

	data, err := cache.GetDict("foo")
	if err != nil {
		t.Fatal("Get key error", err.Error())
	}

	for i, elem := range data {
		if string(elem[2:]) != string(val[i]) {
			t.Fatal("Incorrect element", "expected", string(val[i]), "got", string(elem[2:]))
		}
	}

	indexSearch := map[string]string{
		"zbaz": "world",
		"baz" :"foobaz",
		"fooo" :"baz",
		"c" :"hello",
		"bar" :"BAR",
	}
	for key, val := range indexSearch {
		elem, err := cache.GetDictElement("foo", []byte(key))
		if err != nil {
			t.Fatal("Get key error: ", err.Error())
		}

		if string(elem) != val {
			t.Fatal("Incorrect element", "expected", val, "got", string(elem))
		}
	}

	_, err = cache.GetDictElement("foo", []byte("hello"))
	if err != notFoundErr {
		t.Fatal("Expected error", notFoundErr.Error(), "got", err)
	}
}

func BenchmarkGetParallel(b *testing.B) {
	cache := newSimpleCacheDb()

	cache.Set("foo", 0, []byte("bar"))
	cache.Set("baz", 0, []byte("barbaz"))
	cache.Set("bar", 0, []byte("foobar"))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("foo")
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("baz")
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("bar")
		}
	})
}

func BenchmarkSetParallel(b *testing.B) {
	cache := newSimpleCacheDb()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Set("foo", 0, []byte("bar"))
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Set("baz", 0, []byte("barbaz"))
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Set("bar", 0, []byte("foobar"))
		}
	})
}
