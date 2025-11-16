package dbstorage

import (
	"time"
)

type Retrier interface {
	Retry(f func() error) error
}

type PgRetrier struct {
	retryTimeouts []time.Duration
	classifier    Classifier
}

func NewPgRetrier(retryTimeouts []time.Duration) *PgRetrier {
	return &PgRetrier{
		retryTimeouts: retryTimeouts,
		classifier:    NewPgClassifier(),
	}
}

func (r *PgRetrier) Retry(f func() error) error {
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
