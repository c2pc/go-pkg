package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/retry"
	"github.com/stretchr/testify/assert"
)

func TestRetry_SuccessFirstTry(t *testing.T) {
	ctx := context.Background()

	fn := func() error {
		return nil
	}
	needRetry := func(err error) bool {
		return true
	}

	err := retry.Retry(ctx, fn, needRetry, retry.DefaultDelay, retry.DefaultMaxDelay)
	assert.NoError(t, err)
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()

	attempts := 0
	fn := func() error {
		if attempts < 2 {
			attempts++
			return errors.New("temporary error")
		}
		return nil
	}
	needRetry := func(err error) bool {
		return true
	}

	err := retry.Retry(ctx, fn, needRetry, 10*time.Millisecond, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetry_SuccessAfterRetries2(t *testing.T) {
	ctx := context.Background()

	attempts := 0
	fn := func() error {
		attempts++
		if attempts >= 4 {
			return nil
		}
		return errors.New("temporary error")
	}
	needRetry := func(err error) bool {
		if attempts >= 4 {
			return false
		}
		return true
	}

	err := retry.Retry(ctx, fn, needRetry, 10*time.Millisecond, 1*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, 4, attempts)
}

func TestRetry_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()

	fn := func() error {
		return errors.New("persistent error")
	}
	needRetry := func(err error) bool {
		return true
	}

	err := retry.Retry(ctx, fn, needRetry, 10*time.Millisecond, 50*time.Millisecond)
	assert.ErrorIs(t, err, retry.ErrMaxAttempts)
}

func TestRetry_NeedRetryReturnsFalse(t *testing.T) {
	ctx := context.Background()

	fn := func() error {
		return errors.New("non-retriable error")
	}
	needRetry := func(err error) bool {
		return false
	}

	err := retry.Retry(ctx, fn, needRetry, 10*time.Millisecond, 100*time.Millisecond)
	assert.EqualError(t, err, "non-retriable error")
}

func TestRetry_CanselContext(t *testing.T) {
	ctx, cansel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cansel()

	fn := func() error {
		return errors.New("non-retriable error")
	}
	needRetry := func(err error) bool {
		return true
	}

	err := retry.Retry(ctx, fn, needRetry, 10*time.Millisecond, 1*time.Second)
	assert.EqualError(t, err, retry.ErrContextCansel.Error())
}

func TestRetry_CanselContext2(t *testing.T) {
	ctx, cansel := context.WithCancel(context.Background())
	cansel()

	fn := func() error {
		return errors.New("non-retriable error")
	}
	needRetry := func(err error) bool {
		return true
	}

	err := retry.Retry(ctx, fn, needRetry, 100*time.Millisecond, 1*time.Second)
	assert.EqualError(t, err, retry.ErrContextCansel.Error())
}
