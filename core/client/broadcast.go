package client

import (
	"sync"
)

type broadcaster struct {
	sync.RWMutex
	on   bool
	subs map[chan<- interface{}]chan interface{}
}

type Broadcaster interface {
	Register(chan<- interface{}, int) bool
	Unregister(chan<- interface{})
	Submit(interface{})
}

func NewBroadcaster() Broadcaster {
	b := &broadcaster{
		on:   true,
		subs: make(map[chan<- interface{}]chan interface{}),
	}
	return b
}

func (b *broadcaster) Register(newch chan<- interface{}, buflen int) bool {
	b.Lock()
	defer b.Unlock()
	if !b.on {
		return false
	}
	relayCh := make(chan interface{}, buflen)
	go func() {
		for m := range relayCh {
			newch <- m
		}
	}()
	b.subs[newch] = relayCh
	return true
}

func (b *broadcaster) Unregister(newch chan<- interface{}) {
	b.Lock()
	defer b.Unlock()
	delete(b.subs, newch)
}

// Submit an item to be broadcast to all listeners.
// Message will be missed by listeners if its buffer is full.
func (b *broadcaster) Submit(m interface{}) {
	b.RLock()
	defer b.RUnlock()
	for _, relayCh := range b.subs {
		select {
		case relayCh <- m:
		default:
		}
	}
}
