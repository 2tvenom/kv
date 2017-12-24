package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/2tvenom/kv/client"
)

var ip = flag.String("ip", "127.0.0.1", "kv server ip")
var port = flag.Int("port", 4502, "kv server port")
var number = flag.Int("n", 1000, "request number")
var clients = flag.Int("c", 50, "number of clients")
var round = flag.Int("r", 1, "benchmark round number")
var valueSize = flag.Int("vsize", 100, "kv value size")
var tests = flag.String("t", "set,get,retrand,remove,setlist,getlist,setdict,getdict,getlistelem,getdictelem", "only run the comma separated list of tests")
var wg sync.WaitGroup

var kvClient *client.Client

var loop int = 0

func waitBench(c *client.Client, cmd string) {
	_, err := c.Do(strings.ToUpper(cmd))
	if err != nil {
		fmt.Printf("do %s error %s\n", cmd, err.Error())
	}

}

func bench(cmd string, f func(c *client.Client)) {
	wg.Add(*clients)

	t1 := time.Now()
	for i := 0; i < *clients; i++ {
		go func() {
			for j := 0; j < loop; j++ {
				f(kvClient)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	t2 := time.Now()

	d := t2.Sub(t1)

	fmt.Printf("%s: %s %0.3f micros/op, %0.2fop/s\n",
		cmd,
		d.String(),
		float64(d.Nanoseconds()/1e3)/float64(*number),
		float64(*number)/d.Seconds())
}

var kvSetBase int64 = 0
var kvGetBase int64 = 0
var kvDelBase int64 = 0

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func benchSet() {
	value := randStringBytes(*valueSize)
	f := func(c *client.Client) {
		n := atomic.AddInt64(&kvSetBase, 1)
		waitBench(c, fmt.Sprintf("SET key%d %s", n, value))
	}

	bench("set", f)
}

func benchGet() {
	f := func(c *client.Client) {
		n := atomic.AddInt64(&kvGetBase, 1)
		waitBench(c, fmt.Sprintf("GET key%d", n))
	}

	bench("get", f)
}

func benchRandGet() {
	f := func(c *client.Client) {
		n := rand.Int() % *number
		waitBench(c, fmt.Sprintf("GET key%d", n))
	}

	bench("randget", f)
}

func benchRemove() {
	f := func(c *client.Client) {
		n := atomic.AddInt64(&kvDelBase, 1)
		waitBench(c, fmt.Sprintf("REMOVE %d", n))
	}

	bench("remove", f)
}

func benchSetList() {
	value := ""
	for i := 0; i < 10; i++ {
		if len(value) != 0 {
			value += " "
		}
		value += randStringBytes(10)
	}

	f := func(c *client.Client) {
		waitBench(c, "SETLIST mytestlist "+value)
	}

	bench("setlist", f)
}

func benchGetList() {
	f := func(c *client.Client) {
		waitBench(c, "GETLIST mytestlist")
	}

	bench("getlist", f)
}

func benchGetListElem() {
	f := func(c *client.Client) {
		n := rand.Int() % 10
		waitBench(c, fmt.Sprintf("GETLISTELEM mytestlist %d", n))
	}

	bench("getlistelem", f)
}

func benchSetDict() {
	value := ""
	for i := 0; i < 10; i++ {
		if len(value) != 0 {
			value += " "
		}
		value += string(letterBytes[i]) + ":" + randStringBytes(10)
	}

	f := func(c *client.Client) {
		waitBench(c, "SETDICT mytestdict "+value)
	}

	bench("setdict", f)
}

func benchGetDict() {
	f := func(c *client.Client) {
		waitBench(c, "GETDICT mytestdict")
	}

	bench("getdict", f)
}

func benchGetDictElem() {
	f := func(c *client.Client) {
		n := rand.Int() % 10
		waitBench(c, fmt.Sprintf("GETDICTELEM mytestdict %s", string(letterBytes[n])))
	}

	bench("getdictelem", f)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	flag.Parse()

	if *number <= 0 {
		panic("invalid number")
		return
	}

	if *clients <= 0 || *number < *clients {
		panic("invalid kvClient number")
		return
	}

	loop = *number / *clients

	if *round <= 0 {
		*round = 1
	}

	kvClient = client.NewClient(*ip, *port)
	kvClient.Do("KEYS")


	ts := strings.Split(*tests, ",")

	for i := 0; i < *round; i++ {
		for _, s := range ts {
			switch strings.ToLower(s) {
			case "set":
				benchSet()
			case "get":
				benchGet()
			case "retrand":
				benchRandGet()
			case "remove":
				//benchRemove()
			case "setlist":
				benchSetList()
			case "getlist":
				benchGetList()
			case "getlistelem":
				benchGetListElem()
			case "setdict":
				benchSetDict()
			case "getdict":
				benchGetDict()
			case "getdictelem":
				benchGetDictElem()
			}
		}

		println("")
	}
}
