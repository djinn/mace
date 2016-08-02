package mace

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

type MaceBucket struct {
	sync.RWMutex
	name         string
	items        map[string]*MaceItem
	leakqueue    *leakQueue
	leakTimer    *time.Timer
	leakInterval time.Duration
	logger       *log.Logger
	loadItems    func(string) *MaceItem
	onAddItem    func(*MaceItem)
	onDeleteItem func(*MaceItem)
	accessMax    int
}

func (bucket *MaceBucket) Name() string {
	return bucket.name
}

func (bucket *MaceBucket) Count() int {
	bucket.RLock()
	r := len(bucket.items)
	bucket.RUnlock()
	return r
}

func (bucket *MaceBucket) SetDataLoader(f func(string) *MaceItem) {
	bucket.Lock()
	bucket.loadItems = f
	bucket.Unlock()
}

func (bucket *MaceBucket) SetOnAddItem(f func(*MaceItem)) {
	bucket.Lock()
	bucket.onAddItem = f
	bucket.Unlock()
}

func (bucket *MaceBucket) SetOnDeleteItem(f func(*MaceItem)) {
	bucket.Lock()
	bucket.onDeleteItem = f
	bucket.Unlock()
}

func (bucket *MaceBucket) SetLogger(logger *log.Logger) {
	bucket.Lock()
	bucket.logger = logger
	bucket.Unlock()
}

func (bucket *MaceBucket) leakCheck() {
	bucket.Lock()
	if bucket.leakTimer != nil {
		bucket.leakTimer.Stop()
	}
	if bucket.leakInterval > 0 {
		bucket.log("Expiration check triggered after " + bucket.leakInterval.String() + " for bucket" + bucket.name)
	} else {
		bucket.log("Expiration check installed on bucket", bucket.name)
	}
	invalidL := []*disposeItem{}
	cur := time.Now()
	l := bucket.leakqueue
	for {
		if l.Len() == 0 {
			break
		}
		it := heap.Pop(l)
		if cur.Sub(it.(*disposeItem).disposeTime) <= 0 {
			heap.Push(l, it.(*disposeItem))
			break
		}

		invalidL = append(invalidL, it.(*disposeItem))

	}
	bucket.Unlock()
	// fetch current time for comparison
	// used to create next timer callback

	// Change this to Heap so that cleaning is after
	// at expense of more space usage
	// Per item timestamp + pointer to item
	for _, itemP := range invalidL {
		key := itemP.value
		bucket.delete(key)
	}
	bucket.leakInterval = 0 * time.Millisecond

	bucket.RLock()
	if bucket.leakqueue.Len() > 0 {
		itemMin := heap.Pop(l)
		dur := itemMin.(*disposeItem).disposeTime
		bucket.leakInterval = dur.Sub(cur)
		heap.Push(l, itemMin.(*disposeItem))
		bucket.leakTimer = time.AfterFunc(bucket.leakInterval, func() {
			go bucket.leakCheck()
		})
	}
	bucket.RUnlock()
}

func (bucket *MaceBucket) delete(key string) (*MaceItem, error) {
	bucket.Lock()
	v, ok := bucket.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	v_shadow := NewMaceItem(key, v.data, 0*time.Millisecond)
	delete(bucket.items, key)
	deleteCallback := bucket.onDeleteItem
	bucket.Unlock()

	if deleteCallback != nil {
		// TODO: clone item before calling this routine
		// Secondary advantage is ablility to run this as separate
		// go routine
		deleteCallback(v_shadow)
	}
	return v, nil
}

func (bucket *MaceBucket) Delete(key string) (*MaceItem, error) {
	bucket.Lock()
	v, ok := bucket.items[key]
	if !ok {
		bucket.Unlock()
		return nil, ErrKeyNotFound
	}
	deleteCallback := bucket.onDeleteItem
	bucket.Unlock()
	if deleteCallback != nil {
		// TODO: clone item before calling this routine
		// Secondary advantage is ablility to run this as separate
		// go routine
		deleteCallback(v)
	}
	bucket.Lock()
	bucket.log("Deleting item with key: " + key + " created on " + v.Created().String())
	if v.Alive() != 0 {
		dispose := v.dispose
		if dispose != nil {
			heap.Remove(bucket.leakqueue, dispose.index)
		}
	}
	delete(bucket.items, key)
	bucket.Unlock()
	return v, nil
}

func (bucket *MaceBucket) Set(key string, data interface{},
	alive time.Duration) *MaceItem {
	// If the key already exists
	if bucket.Exists(key) {
		bucket.Delete(key)
	}
	item := NewMaceItem(key, data, alive)
	bucket.Lock()
	bucket.log("Adding item with key: " + key +
		" which will be alive for:" + alive.String())
	bucket.items[key] = item
	if item.alive != 0 {
		heap.Push(bucket.leakqueue, item.dispose)
	}
	expiry := bucket.leakInterval
	addCallback := bucket.onAddItem
	bucket.Unlock()

	if addCallback != nil {
		// TODO: clone item and call addCallback as a go routine
		addCallback(item)
	}
	// Leak check set or run
	if alive > 0 && (expiry == 0 || alive < expiry) {
		bucket.leakCheck()
	}
	return item
}

func (bucket *MaceBucket) Exists(key string) bool {
	bucket.RLock()
	v, ok := bucket.items[key]
	bucket.RUnlock()
	return ok && v != nil
}

func (bucket *MaceBucket) KeepAlive(key string) error {
	bucket.Lock()
	v, ok := bucket.items[key]
	v.KeepAlive()
	// We care to update LeakQueue only if it has Alive duration
	// set
	if ok {
		if v.Alive() != 0 {
			bucket.leakqueue.update(v.dispose)
		}
		bucket.Unlock()
		return nil
	}
	bucket.Unlock()
	return ErrKeyNotFound
}

func (bucket *MaceBucket) Get(key string) (*MaceItem, error) {
	bucket.RLock()
	v, ok := bucket.items[key]
	loadItems := bucket.loadItems
	bucket.RUnlock()
	if ok {
		bucket.KeepAlive(key)
		if bucket.accessMax != 0 && v.AccessCount() >= bucket.accessMax {
			bucket.Delete(key)
		}
		return v, nil
	}
	if loadItems != nil {
		item := loadItems(key)
		if item != nil {
			bucket.Set(key, item.data, item.Alive())
			return item, nil
		}
		return nil, ErrKeyNotFoundOrLoadable
	}
	return nil, ErrKeyNotFound
}

func (bucket *MaceBucket) Flush() {
	bucket.Lock()
	bucket.log("Flushing the cache bucket: " + bucket.name)
	bucket.items = make(map[string]*MaceItem)
	l := leakQueue{}
	heap.Init(&l)
	bucket.leakqueue = &l
	bucket.leakInterval = 0
	if bucket.leakTimer != nil {
		bucket.leakTimer.Stop()
	}
	bucket.Unlock()
	return
}

func (bucket *MaceBucket) log(v ...interface{}) {
	if bucket.logger == nil {
		return
	}
	bucket.logger.Println(v)
}
