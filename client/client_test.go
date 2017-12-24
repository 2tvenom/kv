package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/2tvenom/kv/kv"
	"github.com/2tvenom/kv/server"
)

func TestClient_Do(t *testing.T) {
	addr, port := "127.0.0.1", 4502
	cache := kv.NewCacheDb()

	ts := server.NewTcpServer(cache, addr, port)
	go ts.Listen()

	time.Sleep(time.Second * 2)

	client := NewClient(addr, port)

	val := "value"
	data, err := client.Do("SET key " + val)
	if err != nil {
		t.Fatal("Set error", err.Error())
	}

	out, ok := data.(bool)
	if !ok {
		t.Fatal("Incorrect response", "expected true", "got", out)
	}

	data, err = client.Do("GET key")
	if err != nil {
		t.Fatal("Get error", err.Error())
	}

	outStr, ok := data.(string)
	if !ok {
		t.Fatal("Incorrect response", "expected string", "got", out)
	}

	if outStr != val {
		t.Fatal("Incorrect response", "expected", val, "got", outStr)
	}

	t.Log("String response", outStr)

	_, err = client.Do("SET key")
	if err == nil {
		t.Fatal("Expected error", "got nil")
	}

	data, err = client.Do("SETLIST keylist foo bar baz")
	if err != nil {
		t.Fatal("Setlist error", err.Error())
	}

	out, ok = data.(bool)
	if !ok {
		t.Fatal("Incorrect response", "expected true", "got", out)
	}

	data, err = client.Do("GETLIST keylist")
	if err != nil {
		t.Fatal("Getlist error", err.Error())
	}

	outStrList, ok := data.([]string)
	if !ok {
		t.Fatal("Incorrect response", "expected list", "got", out)
	}

	expectingList := []string{"foo", "bar", "baz"}

	if !reflect.DeepEqual(outStrList, expectingList) {
		t.Fatal("Incorrect response", "expected", expectingList, "got", outStrList)
	}

	t.Log("List response", outStrList)

	data, err = client.Do("SETDICT keydict foo:baz bar:bar baz:foofoo")
	if err != nil {
		t.Fatal("Setdict error", err.Error())
	}

	out, ok = data.(bool)
	if !ok {
		t.Fatal("Incorrect response", "expected true", "got", out)
	}

	data, err = client.Do("GETDICT keydict")
	if err != nil {
		t.Fatal("Getdict error", err.Error())
	}

	outStrDict, ok := data.(map[string]string)
	if !ok {
		t.Fatal("Incorrect response", "expected dictionary", "got", out)
	}

	expectingDict := map[string]string{"foo": "baz", "bar": "bar", "baz": "foofoo"}

	if !reflect.DeepEqual(outStrDict, expectingDict) {
		t.Fatal("Incorrect response", "expected", expectingDict, "got", outStrDict)
	}

	t.Log("Dict response", outStrList)
}
