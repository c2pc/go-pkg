package retry

import (
	"fmt"
	"time"
)

var ErrMaxAttempts = fmt.Errorf("max attempts exceeded error")

var DefaultDelay = 2 * time.Second
var DefaultMaxDelay = 5 * time.Minute

func Retry(fn func() error, needRetry func(error) bool, delay time.Duration, maxDelay time.Duration) error {
	tm := time.Now().Add(maxDelay)
	var attempt int

	for tm.After(time.Now()) {
		err := fn()
		if err == nil {
			return nil
		}
		if needRetry(err) {
			attempt++
			time.Sleep(time.Duration(attempt) * delay)
			continue
		} else {
			return err
		}
	}

	return ErrMaxAttempts
}
