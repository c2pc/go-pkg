package retry

import (
	"context"
	"fmt"
	"time"
)

var ErrMaxAttempts = fmt.Errorf("max attempts exceeded error")
var ErrContextCansel = fmt.Errorf("context canseled")

var DefaultDelay = 2 * time.Second
var DefaultMaxDelay = 5 * time.Minute

func Retry(ctx context.Context, fn func() error, needRetry func(error) bool, delay time.Duration, maxDelay time.Duration) error {
	tm := time.Now().Add(maxDelay)
	var attempt int

	for tm.After(time.Now()) {
		select {
		case <-ctx.Done():
			return ErrContextCansel
		default:
			err := fn()
			if err == nil {
				return nil
			}
			if needRetry(err) {
				attempt++
				select {
				case <-time.After(time.Duration(attempt) * delay):
					continue
				case <-ctx.Done():
					return ErrContextCansel
				}
			} else {
				return err
			}
		}
	}

	return ErrMaxAttempts
}
