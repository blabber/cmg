package lib

import (
	"testing"
	"time"
)

var rateLimiterTests = []struct {
	d            time.Duration
	b            int
	testDuration time.Duration
	expected     int
}{
	{time.Second, 0, time.Second*3 + time.Second/10, 3},
	{time.Second, 5, time.Second*3 + time.Second/10, 8},
	{time.Second * 2, 0, time.Second*3 + time.Second/10, 1},
	{time.Second * 2, 5, time.Second*3 + time.Second/10, 6},
}

func TestRateLimiter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping rate limit tests in short mode")
	}

	for _, tt := range rateLimiterTests {
		endTicker := time.NewTicker(tt.testDuration)
		l := NewRateLimiter(tt.d, tt.b)

		c := 0
		end := make(chan bool)

		go func() {
			for {
				select {
				case <-endTicker.C:
					end <- true
					return
				case <-l.Throttle:
					c = c + 1
				}
			}
		}()

		<-end

		endTicker.Stop()
		l.Stop()

		if c != tt.expected {
			t.Errorf("%d != %d", c, tt.expected)
		}
	}
}

func ExampleRateLimiter() {
	// block for 100 milliseconds, allow bursts of 50 nonblocking calls
	l := NewRateLimiter(time.Second/10, 50)
	defer l.Stop()

	for {
		<-l.Throttle
		// do something that should be rate limited
	}
}
