package ratelimit

import "time"

// TokenBucket implements a token bucket rate limiter.
type TokenBucket struct {
	rate       float64 // Tokens added per second
	capacity   int64   // Maximum tokens in the bucket
	tokens     int64
	lastRefill time.Time

	clock Clock
}

// NewTokenBucket creates a new token bucket with the given parameters.
func NewTokenBucket(rate float64, capacity int64, clock Clock) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity, // Start with a full bucket
		lastRefill: clock.Now(),
		clock:      clock,
	}
}

// Allow checks if a request can be allowed, consuming a token if so.
func (tb *TokenBucket) Allow() bool {
	now := tb.clock.Now()

	// Refill tokens based on elapsed time
	tokensToAdd := int64(now.Sub(tb.lastRefill).Seconds() * tb.rate)
	if tokensToAdd > 0 {
		tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// min returns the smaller of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
