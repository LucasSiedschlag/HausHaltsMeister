package httpapi

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	window time.Duration
	limit  int
	keys   map[string]*rateBucket
}

type rateBucket struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		window: window,
		limit:  limit,
		keys:   make(map[string]*rateBucket),
	}
}

func (r *RateLimiter) Allow(key string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	bucket, ok := r.keys[key]
	if !ok || now.After(bucket.resetTime) {
		r.keys[key] = &rateBucket{count: 1, resetTime: now.Add(r.window)}
		return true, 0
	}

	if bucket.count >= r.limit {
		return false, time.Until(bucket.resetTime)
	}

	bucket.count++
	return true, 0
}
