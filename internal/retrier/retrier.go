package retrier

import (
	"time"
)

type ErrorClass string

var (
	Retriable    ErrorClass = "retriable"
	NonRetriable ErrorClass = "non_retriable"
)

type RetryClassifier interface {
	ClassifyError(err error) ErrorClass
}

type Retrier struct {
	retryTimeouts []time.Duration
	classifier    RetryClassifier
}

func NewRetrier(retryTimeouts []time.Duration, classifier RetryClassifier) *Retrier {
	return &Retrier{
		retryTimeouts: retryTimeouts,
		classifier:    classifier,
	}
}

func (r *Retrier) Retry(f func() error) error {
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
