// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package lib

import "time"

// RateLimiter implements a simple rate limiting mechanism. RateLimiter exports
// a channel named Throttle, which is blocking as specified in the
// NewRateLimiter call.
type RateLimiter struct {
	Throttle <-chan bool
	throttle chan bool
	ticker   *time.Ticker
	running  bool
}

// NewRateLimiter creates a new RateLimiter whose Throttle channel is blocking
// for the time.Duration d before releasing the next bool. If b is greater than
// zero, the Throttle channel is buffered with a buffer size of b, allowing
// bursts of maximal b bools released without blocking. The buffer is initially
// filled, allowing instant bursts.
func NewRateLimiter(d time.Duration, b int) *RateLimiter {
	t := make(chan bool, b)

	l := &RateLimiter{
		t,
		t,
		time.NewTicker(d),
		true,
	}

	// fill throttle initially, allowing instant burst
	for i := 0; i < b; i++ {
		l.throttle <- true
	}

	go func() {
		for l.running {
			<-l.ticker.C
			l.throttle <- true
		}
	}()

	return l
}

// Stop stops the RateLimiter and allows the garbage collector to recover its
// resources.
func (l *RateLimiter) Stop() {
	l.ticker.Stop()
	l.running = false
}
