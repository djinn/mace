package mace

import (
	"sync"
	"time"
)

type MaceItem struct {
	sync.RWMutex
	key         string // Far easier to use string keys
	data        interface{}
	alive       time.Duration
	created     time.Time
	access      time.Time
	accessCount int
	dispose     *disposeItem
	// Expire notification is trait of table structure
	// Likely use case is item specific events are similar in nature
}

func NewMaceItem(key string, val interface{}, aliveUntil time.Duration) *MaceItem {
	cur := time.Now()
	m := &MaceItem{
		key:     key,
		alive:   aliveUntil,
		created: cur,
		access:  cur,
		data:    val,
	}
	if aliveUntil != 0 {
		cur := time.Now()
		disposeTime := cur.Add(aliveUntil)
		ditem := &disposeItem{
			value:       key,
			disposeTime: disposeTime,
		}
		m.dispose = ditem
	}
	return m
}

func (item *MaceItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.access = time.Now()
	item.accessCount++
	if item.alive != 0 {
		item.dispose.disposeTime = item.access.Add(item.alive)
	}
	return
}

func (item *MaceItem) Alive() time.Duration {
	return item.alive
}

func (item *MaceItem) Key() string {
	return item.key
}

func (item *MaceItem) Data() interface{} {
	return item.data
}

func (item *MaceItem) AccessCount() int {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

func (item *MaceItem) Created() time.Time {
	return item.created
}

func (item *MaceItem) Access() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.access
}
