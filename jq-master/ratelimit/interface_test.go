package ratelimit_test

import (
	"testing"

	"github.com/lightpub-dev/lightjq/jq-master/ratelimit"
)

type SimpleLimiter struct {
	allowed bool
}

func (s *SimpleLimiter) Allow() bool {
	return s.allowed
}

func (s *SimpleLimiter) Set(allowed bool) {
	s.allowed = allowed
}

func TestLimiterCollectionBasic(t *testing.T) {
	r1 := &SimpleLimiter{allowed: false}
	r2 := &SimpleLimiter{allowed: true}
	rlc := ratelimit.NewRateLimitterCollection()
	rlc.AddRateLimitter("key1", r1)
	rlc.AddRateLimitter("key2", r2)

	if rlc.Allow("key1") {
		t.Error("Expected key1 to be disallowed")
	}

	if !rlc.Allow("key2") {
		t.Error("Expected key2 to be allowed")
	}
}

func TestLimiterNonExistent(t *testing.T) {
	rlc := ratelimit.NewRateLimitterCollection()

	if !rlc.Allow("unknown") {
		t.Error("unregistered key should be allowed")
	}
}
