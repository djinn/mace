// Simple cache library
// Heavily inspired of github.com/rif/cache2go
// Deviating on finer points
// Copyright (c) 2016, Supreet Sethi <supreet.sethi@gmail.com>
package mace

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type MaceBucket struct {
	sync.RWMutex
	name         string
	items        map[string]*MaceItem
	leakqueue    LeakQueue
	leakTimer    *time.Timer
	leakInterval time.Duration
	logger       *log.Logger
	loadItems    func(string) *MaceItem
	onAddItem    func(*MaceItem)
	onDeleteItem func(*MaceItem)
}

func (bucket *MaceBucket) Count() int {
	bucket.RLock()
	defer bucket.RUnlock()
	return len(bucket.items)
}

func (bucket *MaceBucket) SetDataLoader(f func(string) *MaceItem) {
	bucket.Lock()
	defer bucket.Unlock()
	bucket.loadItems = f
}

func (bucket *MaceBucket) SetOnAddItem(f func(*MaceItem)) {
	bucket.Lock()
	defer bucket.Unlock()
	bucket.onAddItem = f
}

func (bucket *MaceBucket) SetOnDeleteItem(f func(*MaceItem)) {
	bucket.Lock()
	defer bucket.Unlock()
	bucket.onDeleteItem = f
}

func (bucket *MaceBucket) SetLogger(logger *log.Logger) {
	bucket.Lock()
	defer bucket.Unlock()
	bucket.logger = logger
}

func (bucket *MaceBucket) leakCheck() {
	bucket.Lock()
	if bucket.leakTimer != nil {
		bucket.leakTimer.Stop()
	}
	if bucket.leakInterval > 0 {
		bucket.log("Expiration check triggered after" + bucket.leakInterval.String() + "for bucket" + bucket.name)
	} else {
		bucket.log("Expiration check installed on bucket", bucket.name)
	}
	items := bucket.items
	fmt.Printf("%v", bucket.leakqueue)
	bucket.Unlock()

	// fetch current time for comparison
	cur := time.Now()
	// used to create next timer callback

	smallestAlive := 0 * time.Second
	// Change this to Heap so that cleaning is after
	// at expense of more space usage
	// Per item timestamp + pointer to item
	for key, item := range items {
		item.RLock()
		alive := item.Alive()
		access := item.Access()
		item.RUnlock()

		if alive == 0 {
			continue
		}
		if cur.Sub(access) >= alive {
			bucket.Delete(key)
		} else {
			if smallestAlive == 0 || alive < smallestAlive {
				smallestAlive = alive - cur.Sub(access)
			}
		}
		bucket.log("Smallest alive " + smallestAlive.String())
	}
	bucket.Lock()
	bucket.leakInterval = smallestAlive
	if smallestAlive > 0 {
		bucket.leakTimer = time.AfterFunc(smallestAlive, func() {
			go bucket.leakCheck()
		})
	}
	bucket.Unlock()
	return
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
	defer bucket.Unlock()
	bucket.log("Deleting item with key: " + key + " created on " + v.Created().String())
	delete(bucket.items, key)
	return v, nil
}

func (bucket *MaceBucket) Cache(key string, alive time.Duration,
	data interface{}) *MaceItem {
	item := NewMaceItem(key, data, alive)

	bucket.Lock()
	bucket.log("Adding item with key: " + key +
		" which will be alive for:" + alive.String())
	bucket.items[key] = item
	if alive != 0 {
		disposeTime := alive
		ditem := DisposeItem{
			value:       key,
			disposeTime: disposeTime,
		}
		bucket.leakqueue.Push(ditem)
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
	defer bucket.RUnlock()
	_, ok := bucket.items[key]
	return ok
}

func (bucket *MaceBucket) Value(key string) (*MaceItem, error) {
	bucket.RLock()
	v, ok := bucket.items[key]
	loadItems := bucket.loadItems
	bucket.RUnlock()
	if ok {
		v.KeepAlive()
		// We care to update LeakQueue only if it has Alive duration
		// set
		cur := time.Now()
		if v.Alive() != 0 {
			disposeTime := v.Alive() - cur.Sub(v.Access())
			item := &DisposeItem{
				value:       key,
				disposeTime: disposeTime,
			}

			bucket.Lock()
			bucket.leakqueue.Update(item, key, disposeTime)
			bucket.Unlock()
		}
		return v, nil
	}
	if loadItems != nil {
		item := loadItems(key)
		if item != nil {
			bucket.Cache(key, item.Alive(), item.data)
			return item, nil
		}
		return nil, ErrKeyNotFoundOrLoadable
	}
	return nil, ErrKeyNotFound
}

func (bucket *MaceBucket) Flush() {
	bucket.Lock()
	defer bucket.Unlock()
	bucket.log("Flushing the cache bucket: " + bucket.name)
	bucket.items = make(map[string]*MaceItem)
	bucket.leakqueue = LeakQueue{}
	bucket.leakInterval = 0
	if bucket.leakTimer != nil {
		bucket.leakTimer.Stop()
	}
	return
}

func (bucket *MaceBucket) log(v ...interface{}) {
	if bucket.logger == nil {
		return
	}
	bucket.logger.Println(v)
}
