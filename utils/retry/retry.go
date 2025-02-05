package retry

import (
	"errors"
	"fmt"
	"time"
)

var ErrMaxAttempts = fmt.Errorf("max attempts exceeded error")

var DefaultDelays = []time.Duration{2 * time.Second}

func Retry(fn func() error, needRetry func(error) bool, delay time.Duration, maxAttempts int) error {
	for attempts := 0; attempts < maxAttempts; attempts++ {
		err := fn()
		if err == nil {
			return nil
		}
		if needRetry(err) {
			if attempts < maxAttempts {
				time.Sleep(delay)
				continue
			}
			return errors.Join(ErrMaxAttempts, err)
		} else {
			return err
		}
	}

	return ErrMaxAttempts
}
