// Simple cache library
// Heavily inspired of github.com/rif/cache2go
// Deviating on finer points
// Copyright (c) 2016, Supreet Sethi <supreet.sethi@gmail.com>

package mace

import (
	"container/heap"
	"sync"
)

var (
	mace  = make(map[string]*MaceBucket)
	mutex sync.RWMutex
)

func Mace(bucket_name string) *MaceBucket {
	mutex.RLock()
	b, ok := mace[bucket_name]
	mutex.RUnlock()
	l := LeakQueue{}
	heap.Init(&l)
	if !ok {
		b = &MaceBucket{
			name:      bucket_name,
			items:     make(map[string]*MaceItem),
			leakqueue: &l,
		}
		mutex.Lock()
		mace[bucket_name] = b
		mutex.Unlock()
	}
	return b
}
