package failsafe

import (
	"context"
	"errors"
)

var (
	ErrContextCancelled = errors.New("context cancelled")
)

type Failsafe struct {
	retryPolicy *RetryPolicy
}

func New(retryPolicy *RetryPolicy) *Failsafe {
	return &Failsafe{
		retryPolicy: retryPolicy,
	}
}

func (f *Failsafe) Run(ctx context.Context, attempt func() error) error {
	var err error

	for !f.retryPolicy.IsDone() {
		select {
		case <-ctx.Done():
			f.retryPolicy.Cancel()
			err = ErrContextCancelled
		case <-f.retryPolicy.Next():
			err = attempt()
			if err == nil {
				return nil
			}

			f.retryPolicy.Report(err)
		}
	}

	return err
}
