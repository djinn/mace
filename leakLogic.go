// Simple cache library
// Heavily inspired of github.com/rif/cache2go
// Deviating on finer points
// Copyright (c) 2016, Supreet Sethi <supreet.sethi@gmail.com>

package mace

import (
	"container/heap"
	"fmt"
	"time"
)

type DisposeItem struct {
	value       string
	disposeTime time.Duration
	index       int
}

type LeakQueue []*DisposeItem

func (lq LeakQueue) Len() int {
	return len(lq)
}

func (lq LeakQueue) Less(i, j int) bool {
	return lq[i].disposeTime < lq[j].disposeTime
}

func (lq LeakQueue) Swap(i, j int) {
	lq[i], lq[j] = lq[j], lq[i]
	return
}

func (lq *LeakQueue) Push(key interface{}) {
	n := len(*lq)
	item := key.(DisposeItem)
	item.index = n
	*lq = append(*lq, &item)
	return
}

func (lq *LeakQueue) String() string {
	return fmt.Sprintf("Leakqueue len: %d", lq.Len())
}

func (lq *LeakQueue) Pop() interface{} {
	old := *lq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*lq = old[0 : n-1]
	return item
}

func (lq *LeakQueue) Update(item *DisposeItem, value string, disposeTime time.Duration) {
	item.value = value
	item.disposeTime = disposeTime
	heap.Fix(lq, item.index)
}
