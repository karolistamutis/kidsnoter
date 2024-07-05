package util

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func RetryWithBackoff(ctx context.Context, operation func() error) error {
	base, cap := time.Second, time.Minute
	maxAttempts := 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		if attempt == maxAttempts-1 {
			return err
		}

		backoff := float64(base) * math.Pow(2, float64(attempt))
		if backoff > float64(cap) {
			backoff = float64(cap)
		}

		jitter := rand.Float64() * backoff * 0.1
		backoff += jitter

		timer := time.NewTimer(time.Duration(backoff))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retry attempts reached")
}
