package ratelimit

import "time"

// Used to get the current time for rate limiting.
type Clock interface {
	Now() time.Time
}

var (
	DefaultClock Clock = &RealClock{}
)

// RealClock is a real-time clock.
type RealClock struct{}

func (r *RealClock) Now() time.Time {
	return time.Now()
}

// Generic rate limiter interface for various implementations.
type RateLimiter interface {
	// Allow checks if a request can be allowed, consuming a token if so.
	Allow() bool
}

type RateLimiterCollection struct {
	// Rate limiters for different keys
	limitters map[string]RateLimiter
}

// NewRateLimitterCollection creates a new rate limiter collection.
func NewRateLimitterCollection() *RateLimiterCollection {
	return &RateLimiterCollection{
		limitters: make(map[string]RateLimiter),
	}
}

// AddRateLimitter adds a rate limiter for a specific key.
func (rlc *RateLimiterCollection) AddRateLimitter(key string, rl RateLimiter) {
	rlc.limitters[key] = rl
}

func (rlc *RateLimiterCollection) Allow(key string) bool {
	rl, ok := rlc.limitters[key]
	if !ok {
		return true
	}
	return rl.Allow()
}
