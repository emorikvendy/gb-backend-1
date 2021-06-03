package main

import (
	"log"
	"net"
	"sync"
)

type ConnSet struct {
	mx sync.RWMutex
	m  map[net.Conn]struct{}
}

func NewConnSet() *ConnSet {
	return &ConnSet{
		m: map[net.Conn]struct{}{},
	}
}
func (set *ConnSet) Get(key net.Conn) (struct{}, bool) {
	set.mx.RLock()
	defer set.mx.RUnlock()
	val, ok := set.m[key]
	return val, ok
}

func (set *ConnSet) Has(key net.Conn) bool {
	set.mx.RLock()
	defer set.mx.RUnlock()
	_, ok := set.m[key]
	return ok
}

func (set *ConnSet) Add(key net.Conn, value struct{}) {
	set.mx.Lock()
	defer set.mx.Unlock()
	set.m[key] = value
}
func (set *ConnSet) Delete(key net.Conn) {
	set.mx.Lock()
	defer set.mx.Unlock()
	if _, ok := set.m[key]; ok {
		delete(set.m, key)
	}
}

func (set *ConnSet) Range(f func(key net.Conn)) {
	set.mx.Lock()
	defer set.mx.Unlock()
	for conn := range set.m {
		log.Printf("started range function for connection %v", conn)
		f(conn)
		log.Printf("ended range function for connection %v", conn)
	}
}
