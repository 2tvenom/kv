package main

import (
	"testing"
)

func TestCmdParserKeys(t *testing.T) {
	cmdTestCase1 := map[string]string{
		"GET key":      "key",
		"GETLIST key2": "key2",
		"GETDICT key3": "key3",
		"REMOVE key6":  "key6",
	}

	for testData, result := range cmdTestCase1 {
		parser := &baseCommandParser{}
		_, err := parser.Write([]byte(testData))
		if err != nil {
			t.Fatal("Got Error:", err)
		}

		if result != parser.key {
			t.Fatal("Incorrect key", "expected", result, "got", parser.key)
		}
	}

	cmdTestCase2 := map[string]string{
		"GET key aaaa":      "key",
		"GETLIST key2 bbbb": "key2",
		"GETDICT key3 cccc": "key3",
		"REMOVE key6 dddd":  "key6",
		"KEYS key7":         "",
	}

	for testData, result := range cmdTestCase2 {
		parser := &baseCommandParser{}
		_, err := parser.Write([]byte(testData))
		if err != nil {
			t.Fatal("Got Error:", err)
		}

		if result != parser.key {
			t.Fatal("Incorrect key", "expected", result, "got", parser.key)
		}

		if nil != parser.value {
			t.Fatal("Incorrect value", "expected null", "got", parser.value)
		}
	}
}

func TestCmdFullParser(t *testing.T) {
	type (
		testCase struct {
			in      string
			cmd     string
			key     string
			ttl     int64
			value   string
			isError bool
		}
	)

	testCases := []*testCase{
		{"GET aa", "GET", "aa", 0, "", false},
		{"GET zz xx", "GET", "zz", 0, "", false},
		{"GET zz 1 xx", "GET", "zz", 0, "", false},
		{"SET aa 1 ", "SET", "aa", 1, "", true},
		{"SET kk 798ds aaa", "SET", "kk", 0, "798ds aaa", false},
		{"SET bb 1", "SET", "bb", 0, "1", false},
		{"SETsssbb 1", "SET", "bb", 0, "1", true},
		{"SET sssbb 800 hello", "SET", "sssbb", 800, "hello", false},
		{"SETDICT ee a:1 b:2", "SETDICT", "ee", 0, "a:1 b:2", false},
		{"SETDICT ee 44 a:1 b:2", "SETDICT", "ee", 44, "a:1 b:2", false},
		{"SETLIST ff a b c", "SETLIST", "ff", 0, "a b c", false},
		{"SETLIST ff 22 a b c", "SETLIST", "ff", 22, "a b c", false},
		{"KEYS", "KEYS", "", 0, "", false},
		{"REMOVE", "REMOVE", "", 0, "", true},
		{"REMOVE hhh", "REMOVE", "hhh", 0, "", false},
	}

	for _, tc := range testCases {
		t.Log(tc.in)

		parser := &baseCommandParser{}
		_, err := parser.Write([]byte(tc.in))

		if tc.isError {
			if err == nil {
				t.Fatal("Expected Error", "got nil")
			}
			continue
		}

		if err != nil {
			t.Fatal("Got Error:", err)
		}

		if parser.cmd != tc.cmd {
			t.Fatal("Incorrect cmd", "expected", tc.cmd, "got", parser.cmd)
		}

		if parser.key != tc.key {
			t.Fatal("Incorrect key", "expected", tc.key, "got", parser.key)
		}

		if parser.ttl != tc.ttl {
			t.Fatal("Incorrect ttl", "expected", tc.ttl, "got", parser.ttl)
		}

		if string(parser.value) != tc.value {
			t.Fatal("Incorrect value", "expected", tc.value, "got", string(parser.value))
		}
	}
}

func TestCmdParserTTL(t *testing.T) {
	type (
		testCase struct {
			in     string
			ttl    int64
			offset int
			err    error
		}
	)

	testCases := []*testCase{
		{"1 ", 1, 2, nil},
		{"10 ", 10, 3, nil},
		{"1a0 ", 0, 0, notTTl},
		{"10", 0, 0, notTTl},
		{"1543534534543543 ", 1543534534543543, 17, nil},
		{"1543534534543543", 0, 0, notTTl},
		{"0 ", 0, 0, zeroTTl},
	}

	for _, tc := range testCases {
		ttl, offset, err := parseTTL([]byte(tc.in))
		if err != tc.err {
			t.Fatal("Incorrect Error", "expected", tc.err, "got", err)
		}

		if offset != tc.offset {
			t.Fatal("Incorrect offset", "expected", tc.offset, "got", offset)
		}

		if ttl != tc.ttl {
			t.Fatal("Incorrect ttl", "expected", tc.ttl, "got", ttl)
		}
	}
}

func TestCmdParserBigTTL(t *testing.T) {
	_, _, err := parseTTL([]byte("79874892749237947293748923432423423 "))
	if err == nil {
		t.Fatal("Expected Error, got nil")
	}
}
