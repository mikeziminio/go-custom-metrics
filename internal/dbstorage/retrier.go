package dbstorage

import (
	"time"
)

type retrier interface {
	Retry(f func() error) error
}

type pgRetrier struct {
	retryTimeouts []time.Duration
	classifier    classifier
}

func NewPgRetrier(retryTimeouts []time.Duration) *pgRetrier {
	return &pgRetrier{
		retryTimeouts: retryTimeouts,
		classifier:    NewPgClassifier(),
	}
}

func (r *pgRetrier) Retry(f func() error) error {
	var timeoutIndex int
	for {
		e := f()
		if timeoutIndex >= len(r.retryTimeouts) {
			return e
		}
		if class := r.classifier.ClassifyError(e); class == NonRetriable {
			return e
		}
		time.Sleep(r.retryTimeouts[timeoutIndex])
		timeoutIndex++
	}
}
