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
	end      chan<- bool
}

// NewRateLimiter creates a new RateLimiter whose Throttle channel is blocking
// for the time.Duration d before releasing the next bool. If b is greater than
// zero, the Throttle channel is buffered with a buffer size of b, allowing
// bursts of maximal b bools released without blocking. The buffer is initially
// filled, allowing instant bursts.
func NewRateLimiter(d time.Duration, b int) RateLimiter {
	throttle := make(chan bool, b)
	end := make(chan bool)
	ticker := time.NewTicker(d)

	// fill throttle initially, allowing instant burst
	for i := 0; i < b; i++ {
		throttle <- true
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				throttle <- true
			case <-end:
				ticker.Stop()
				return
			}
		}
	}()

	return RateLimiter{
		throttle,
		end,
	}
}

// Stop stops the RateLimiter and allows the garbage collector to recover its
// resources.
func (l RateLimiter) Stop() {
	l.end <- true
}
