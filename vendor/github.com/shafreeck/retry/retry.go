package retry

import (
	"context"
	"time"
)

// RetriableErr is an error type which can be retried
type RetriableErr struct {
	error
}

// Retriable makes an error be retriable
func Retriable(err error) *RetriableErr {
	return &RetriableErr{err}
}

// Retry ensures that the do function will be executed until some condition being satisfied
type Retry struct {
	backoff BackoffStrategy
	base    time.Duration
}

var r = New()

// Ensure keeps retring until ctx is done
func (r *Retry) Ensure(ctx context.Context, do func() error) error {
	duration := r.base
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := do(); err != nil {
			if _, ok := err.(*RetriableErr); ok {
				if r.backoff != nil {
					duration = r.backoff(duration)

					time.Sleep(duration)
				}
				continue
			}
			return err
		}

		return nil
	}
}

// Option is an option to new a Retry object
type Option func(r *Retry)

// BackoffStrategy defines the backoff strategy of retry
type BackoffStrategy func(last time.Duration) time.Duration

// WithBackoff replace the default backoff function
func WithBackoff(backoff BackoffStrategy) Option {
	return func(r *Retry) {
		r.backoff = backoff
	}
}

// WithBaseDelay set the first delay duration, default 10ms
func WithBaseDelay(base time.Duration) Option {
	return func(r *Retry) {
		r.base = base
	}
}

// New a retry object
func New(opts ...Option) *Retry {
	r := &Retry{base: 10 * time.Millisecond, backoff: Exponential(2)}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Exponential generates backoff duration by expoential
func Exponential(factor float64) BackoffStrategy {
	return func(last time.Duration) time.Duration {
		return time.Duration(float64(last) * factor)
	}
}

// Tick keeps a constant backoff interval
func Tick(tick time.Duration) BackoffStrategy {
	return func(last time.Duration) time.Duration {
		return tick
	}
}

// Ensure keeps retring until ctx is done, it use a default retry object
func Ensure(ctx context.Context, do func() error) error {
	return r.Ensure(ctx, do)
}
