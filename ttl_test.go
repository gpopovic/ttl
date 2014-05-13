package ttl

import (
	"testing"
	"time"
)

func TestGeneral(t *testing.T) {
	cache := New(time.Second)
	cache.Add("item", "1")

	// Get before expires
	_, ok := cache.Get("item")
	if !ok {
		t.Error("item key should exist, but doesn't")
	}

	go func() {
		cache.Duration = time.Second * 2
		cache.Add("next", "2")

		_, ok := cache.Get("next")
		if !ok {
			t.Error("next key should exist, but doesn't")
		}
	}()

	// Wait for items to expire
	<-time.After(time.Second * 2)
	_, ok = cache.Get("item")
	if ok {
		t.Error("item key exists, but shouldn't")
	}

	<-time.After(time.Second)
	_, ok = cache.Get("next")
	if ok {
		t.Error("next key exists, but shouldn't")
	}
}

func TestResetOnAdd(t *testing.T) {
	cache := New(time.Second * 2)
	cache.ResetOnAdd = true
	cache.Add("item", "1")

	<-time.After(time.Second)
	cache.Add("item", "2")

	val, ok := cache.Get("item")
	if !ok {
		t.Error("item key doesn't exist, but should")
	}
	if val != "2" {
		t.Error("item key not the expected value. Expected: 2, Actual:", val)
	}

	<-time.After(time.Second * 2)
	_, ok = cache.Get("item")
	if ok {
		t.Error("item key exists, but shouldn't")
	}
}
