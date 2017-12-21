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

	_, err := cache.Get("foo")
	if err == nil {
		t.Fatal("Expected error", "got nil")
	}

	if err != notFoundErr {
		t.Fatal("Incorrect error", "expected", notFoundErr, "got", err)
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