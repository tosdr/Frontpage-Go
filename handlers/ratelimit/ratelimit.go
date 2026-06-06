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
	buckets    map[string]*TokenBucket
	maxTokens  float64
	refillRate float64
	mu         sync.RWMutex
}

var Limiter = NewRateLimiter()

// ContactLimiter is a strict per-IP limiter for the contact form: a burst of 3
// submissions, then 1 token refilled per minute. A real user submits the form
// rarely, so this stops automated spam floods while staying invisible to humans.
var ContactLimiter = NewRateLimiterWith(3, 1.0/60.0)

func NewRateLimiter() *RateLimiter {
	return NewRateLimiterWith(5, 0.2) // 5 tokens max, refill 1 token every 5 seconds
}

func NewRateLimiterWith(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		maxTokens:  maxTokens,
		refillRate: refillRate,
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
		bucket = newTokenBucket(r.maxTokens, r.refillRate)
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
