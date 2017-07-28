package failsafe

import (
	"time"

	"github.com/Sirupsen/logrus"
)

type RetryPolicy struct {
	Spec RetryPolicySpec

	ticker      *time.Ticker
	retriesLeft int
}

type RetryPolicySpec struct {
	Delay    time.Duration
	MaxRetry int
}

func NewRetryPolicy() *RetryPolicy {
	spec := RetryPolicySpec{
		Delay:    1 * time.Nanosecond,
		MaxRetry: 60,
	}

	retryPolicy := &RetryPolicy{
		Spec:        spec,
		ticker:      time.NewTicker(spec.Delay),
		retriesLeft: spec.MaxRetry,
	}

	return retryPolicy
}

func (r *RetryPolicy) IsDone() bool {
	return r.retriesLeft == 0
}

func (r *RetryPolicy) Cancel() {
	r.retriesLeft = 0
	r.ticker.Stop()
}

func (r *RetryPolicy) Next() <-chan time.Time {
	return r.ticker.C
}

func (r *RetryPolicy) Report(err error) {
	logrus.Debugf("Received error: %s", err)
}

func (r *RetryPolicy) WithDelay(delay time.Duration) *RetryPolicy {
	r.Spec.Delay = delay
	r.ticker.Stop()
	r.ticker = time.NewTicker(delay)
	return r
}
