package main

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int
	refillRate float64
	tokens     float64
	lastRefill time.Time

	mu sync.Mutex
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     float64(capacity),
		lastRefill: time.Now(),
	}
}

func (t *TokenBucket) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	duration := now.Sub(t.lastRefill).Seconds()
	refill := duration * t.refillRate
	t.tokens = t.tokens + refill
	t.lastRefill = now

	if t.tokens > float64(t.capacity) {
		t.tokens = float64(t.capacity)
	}

	if t.tokens >= 1 {
		t.tokens--
		return true
	}
	return false
}
