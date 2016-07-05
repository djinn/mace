// Simple cache library
// Heavily inspired of github.com/rif/cache2go
// Deviating on finer points
// Copyright (c) 2016, Supreet Sethi <supreet.sethi@gmail.com>

package mace

import (
	"fmt"
	"time"
)

type disposeItem struct {
	value       string
	disposeTime time.Time
	index       int
}

type leakQueue []*disposeItem

func (lq leakQueue) Len() int {
	return len(lq)
}

func (lq leakQueue) Less(i, j int) bool {
	if lq[i].disposeTime.Sub(lq[j].disposeTime) <= 0 {
		return true
	}
	return false
}

func (lq leakQueue) Swap(i, j int) {
	lq[i], lq[j] = lq[j], lq[i]
	return
}

func (lq leakQueue) String() string {
	var s string
	for _, i := range lq {
		s = s + fmt.Sprintf("Value: %s, disposeTime %s\n", i.value, i.disposeTime)
	}
	return s
}

func (lq *leakQueue) Push(key interface{}) {
	n := len(*lq)
	item := key.(*disposeItem)
	item.index = n
	*lq = append(*lq, item)
	return
}

func (lq *leakQueue) Pop() interface{} {
	old := *lq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*lq = old[0 : n-1]
	return item
}
