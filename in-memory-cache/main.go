package main

import (
	"sync"
	"time"
)

// REQUIREMENTS
//  - Store key → value. Get returns the value and a found flag.
//  - Every value expires after a TTL (time to live).
//  - Safe to use from many goroutines at the same time.
//  - (mid) Maximum size. When full, remove the Least Recently Used entry.
//  - Out of scope: persistence, multiple servers.

type item struct {
	val       string
	expiresAt time.Time
}

type cache struct {
	items map[string]item
	ttl   time.Duration
	mu    sync.Mutex
}

func NewSimpleCache(ttl time.Duration) *cache {
	return &cache{
		items: make(map[string]item),
		ttl:   ttl,
	}
}

func (c *cache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = item{
		val:       v,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *cache) Get(k string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[k]
	if !ok || time.Now().After(it.expiresAt) {
		delete(c.items, k) // lazy expire
		return "", false
	}
	return it.val, true
}

func (c *cache) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, k)
}
