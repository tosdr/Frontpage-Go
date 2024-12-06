package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
	mu             sync.Mutex
}

type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

var Limiter = NewRateLimiter()

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
}

func newTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	bucket, exists := r.buckets[key]
	if !exists {
		bucket = newTokenBucket(5, 0.2) // 5 tokens max, refill 1 token every 5 seconds
		r.buckets[key] = bucket
	}
	r.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	timePassed := now.Sub(bucket.lastRefillTime).Seconds()
	bucket.tokens = min(bucket.maxTokens, bucket.tokens+timePassed*bucket.refillRate)
	bucket.lastRefillTime = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
