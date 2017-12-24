package kv

import "bytes"

type (
	dictionary [][]byte
)

var (
	dictionarySeparator = []byte(":")
)

func (s dictionary) Len() int {
	return len(s)
}
func (s dictionary) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s dictionary) Less(i, j int) bool {
	return bytes.Compare(s[i][0:bytes.Index(s[i], dictionarySeparator)], s[j][0:bytes.Index(s[j], dictionarySeparator)]) < 0
}
