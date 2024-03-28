package ratelimit_test

import (
	"testing"
	"time"

	"github.com/lightpub-dev/lightjq/jq-master/ratelimit"
)

type TestableClock struct {
	now time.Time
}

func (c *TestableClock) Now() time.Time {
	return c.now
}

func (c *TestableClock) Add(d time.Duration) {
	c.now = c.now.Add(d)
}

func TestBasicAllowance(t *testing.T) {
	clk := &TestableClock{now: time.Now()}
	limiter := ratelimit.NewTokenBucket(2, 2, clk) // 2 tokens per second, capacity of 2

	ok1 := limiter.Allow()
	ok2 := limiter.Allow()
	ok3 := limiter.Allow()
	if !ok1 || !ok2 || ok3 {
		t.Error("Incorrect allowance behavior in basic case")
	}
}

func TestRefilling(t *testing.T) {
	clk := &TestableClock{now: time.Now()}
	limiter := ratelimit.NewTokenBucket(1, 3, clk) // 1 token per second, capacity of 3

	ok1 := limiter.Allow()
	ok2 := limiter.Allow()
	ok3 := limiter.Allow()
	ok4 := limiter.Allow()
	if !ok1 || !ok2 || !ok3 || ok4 {
		t.Error("Incorrect allowance behavior in refilling case")
	}

	clk.Add(1500 * time.Millisecond)
	ok5 := limiter.Allow()
	ok6 := limiter.Allow()
	if !ok5 || ok6 {
		t.Error("Incorrect allowance behavior after refill")
	}
}
