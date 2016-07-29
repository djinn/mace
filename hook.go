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

func MaceAccessMax(bucket_name string, accessMax int) *MaceBucket {
	mutex.RLock()
	b, ok := mace[bucket_name]
	mutex.RUnlock()
	if !ok {
		l := leakQueue{}
		heap.Init(&l)
		b = &MaceBucket{
			name:      bucket_name,
			items:     make(map[string]*MaceItem),
			leakqueue: &l,
			accessMax: accessMax,
		}
		mutex.Lock()
		mace[bucket_name] = b
		mutex.Unlock()
	}
	return b

}

func Mace(bucket_name string) *MaceBucket {
	return MaceAccessMax(bucket_name, 0)
}
