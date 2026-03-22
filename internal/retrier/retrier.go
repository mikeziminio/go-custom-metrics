// Package retrier provides retry logic for operations with configurable timeouts.
//
// It supports classifying errors into retriable and non-retriable categories
// and automatically retries operations based on a configurable timeout schedule.
package retrier

import (
	"time"
)

// ErrorClass represents the classification of an error as retriable or non-retriable.
type ErrorClass string

var (
	// Retriable indicates the operation should be retried.
	Retriable ErrorClass = "retriable"
	// NonRetriable indicates the operation should not be retried.
	NonRetriable ErrorClass = "non_retriable"
)

// RetryClassifier interface classifies errors as retriable or non-retriable.
type RetryClassifier interface {
	ClassifyError(err error) ErrorClass
}

// Retrier manages retry logic for operations with configurable timeouts.
type Retrier struct {
	retryTimeouts []time.Duration
	classifier    RetryClassifier
}

// NewRetrier creates a new Retrier instance with the specified timeout schedule.
//
// Parameters:
//   - retryTimeouts: Slice of durations to wait between retries
//   - classifier: RetryClassifier implementation for error classification
//
// Returns a new *Retrier ready to use.
func NewRetrier(retryTimeouts []time.Duration, classifier RetryClassifier) *Retrier {
	return &Retrier{
		retryTimeouts: retryTimeouts,
		classifier:    classifier,
	}
}

// Retry executes a function with automatic retry logic.
//
// Parameters:
//   - f: Function to execute and potentially retry
//
// Returns the error from the last execution of f, or nil if successful.
// The function will be retried based on the configured timeout schedule
// and error classification until either success or all retries are exhausted.
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
