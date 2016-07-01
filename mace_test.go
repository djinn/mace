// Simple cache library
// Heavily inspired of github.com/rif/cache2go
// Deviating on finer points
// Copyright (c) 2016, Supreet Sethi <supreet.sethi@gmail.com>
package mace

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	k      = "macekey"
	v      = "macevalue"
	logger = log.New(os.Stdout, "Mace:", log.LstdFlags)
)

func TestMaceCache(t *testing.T) {
	bucket := Mace("testMace")
	bucket.Cache(k, 1*time.Second, v)
	p, err := bucket.Value(k)
	if err != nil || p == nil || p.Data().(string) != v {
		t.Error("Error retrieving data from cache", err)
	}
}

func TestMaceCacheExpire(t *testing.T) {
	bucket := Mace("testMaceExpire")
	bucket.Cache(k, 250*time.Millisecond, v)
	p, err := bucket.Value(k)
	if err != nil || p == nil || p.Data().(string) != v {
		t.Error("Error retrieving data from cache", err)
	}
	time.Sleep(500 * time.Millisecond)
	p, err = bucket.Value(k)
	if err == nil || p != nil {
		t.Errorf("%v %s\n", p, err)
		t.Error("Error expiring data")
	}
}

func TestMaceCacheNonExpiring(t *testing.T) {
	bucket := Mace("testMaceNonExpiring")
	bucket.Cache(k, 0, v)
	time.Sleep(500 * time.Millisecond)
	p, err := bucket.Value(k)
	if err != nil || p == nil || p.Data().(string) != v {
		t.Error("Error retrieving data from cache", err)
	}
}

func TestMaceCacheKeepAlive(t *testing.T) {
	k2 := k + k
	v2 := v + v
	bucket := Mace("testMaceKeepAlive")
	bucket.Cache(k, 250*time.Millisecond, v)
	bucket.Cache(k2, 750*time.Millisecond, v2)

	p, err := bucket.Value(k)
	if err != nil || p == nil || p.Data().(string) != v {
		t.Error("Error retrieving data from cache", err)
	}
	time.Sleep(50 * time.Millisecond)
	p.KeepAlive()

	time.Sleep(450 * time.Millisecond)
	p, err = bucket.Value(k)
	if err == nil || p != nil {
		t.Error("Error expiring data")
	}
	p, err = bucket.Value(k2)
	if err != nil || p == nil || p.Data().(string) != v2 {
		t.Error("Error retrieving data from cache", err)
	}
	time.Sleep(1 * time.Second)
	p, err = bucket.Value(k2)
	if err == nil || p != nil {
		t.Error("Error expiring data")
	}
}

func TestMaceExists(t *testing.T) {
	bucket := Mace("testMaceExists")
	bucket.Cache(k, 0, v)
	if !bucket.Exists(k) {
		t.Error("Error verifying existing data in cache")
	}
}

func TestMaceDelete(t *testing.T) {
	bucket := Mace("testMaceDelete")
	bucket.Cache(k, 0, v)
	p, err := bucket.Value(k)
	if err != nil || p == nil || p.Data().(string) != v {
		t.Error("Error retrieving data from cache", err)
	}
	bucket.Delete(k)
	p, err = bucket.Value(k)
	if err == nil || p != nil {
		t.Error("Error deleting data")
	}
}

func TestMaceFlush(t *testing.T) {
	bucket := Mace("testMaceFlush")
	bucket.Cache(k, 10*time.Second, v)
	time.Sleep(100 * time.Millisecond)
	bucket.Flush()

	p, err := bucket.Value(k)
	if err == nil || p != nil {
		t.Error("Error expiring data")
	}
	if bucket.Count() != 0 {
		t.Error("Error verifying empty bucket")
	}
}

func TestMaceFlushNoTimout(t *testing.T) {
	bucket := Mace("testMaceFlushNoTimeout")
	bucket.Cache(k, 10*time.Second, v)
	bucket.Flush()

	p, err := bucket.Value(k)
	if err == nil || p != nil {
		t.Error("Error expiring data")
	}
	if bucket.Count() != 0 {
		t.Error("Error verifying empty bucket")
	}
}

func TestMaceCount(t *testing.T) {
	count := 100000
	bucket := Mace("testCount")
	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		bucket.Cache(key, 10*time.Second, v)
	}
	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		p, err := bucket.Value(key)
		if err != nil || p == nil || p.Data().(string) != v {
			t.Error("Error retrieving data")
		}
	}
	if bucket.Count() != count {
		t.Error("Data count mismatch")
	}
}

func TestMaceDataLoader(t *testing.T) {
	bucket := Mace("testMaceDataLoader")
	bucket.SetDataLoader(func(key string) *MaceItem {
		var item *MaceItem
		if key != "nil" {
			val := key
			i := NewMaceItem(key, val, 500*time.Millisecond)
			item = i
		}
		return item
	})

	p, err := bucket.Value("nil")
	if err == nil || bucket.Exists("nil") {
		t.Error("Error validating data loader for nil values")
	}

	for i := 0; i < 10; i++ {
		key := k + strconv.Itoa(i)
		vp := key
		p, err = bucket.Value(key)
		if err != nil || p == nil || p.Data().(string) != vp {
			t.Error("Error validating data loader")
		}
	}
}

func TestMaceCallbacks(t *testing.T) {
	addedKey := ""
	removedKey := ""

	bucket := Mace("testMaceCallbacks")
	bucket.SetOnAddItem(func(item *MaceItem) {
		addedKey = item.Key()
	})
	bucket.SetOnDeleteItem(func(item *MaceItem) {
		removedKey = item.Key()
	})

	bucket.Cache(k, 500*time.Millisecond, v)

	time.Sleep(250 * time.Millisecond)
	if addedKey != k {
		t.Error("AddedItem callback not working")
	}

	time.Sleep(500 * time.Millisecond)
	if removedKey != k {
		t.Error("AboutToDeleteItem callback not working:" + k + "_" + removedKey)
	}
}
