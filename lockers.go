package main

import "sync"

type (
	lockers struct {
		locks  [blocks]sync.RWMutex
	}
)

func (l *lockers) Lock(block uint8) {
	l.locks[block].Lock()
}

func (l *lockers) Unlock(block uint8) {
	l.locks[block].Unlock()
}

func (l *lockers) RLock(block uint8) {
	l.locks[block].RLock()
}

func (l *lockers) RUnlock(block uint8) {
	l.locks[block].RUnlock()
}
