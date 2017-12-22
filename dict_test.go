package main

import (
	"sort"
	"testing"
)

func TestDictionarySort(t *testing.T) {
	dictionary := dictionary{
		[]byte("foo:baz"),
		[]byte("fooo:baz"),
		[]byte("baz:foobaz"),
		[]byte("zbaz:world"),
		[]byte("bar:BAR"),
		[]byte("c:hello"),
		[]byte("d:hello"),
	}
	sort.Sort(dictionary)

	for _, data := range dictionary {
		t.Log(string(data))
	}

}
