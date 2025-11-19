package dbstorage

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPgRetrier(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		retrier := NewPgRetrier([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond})

		attempts := 0
		err := retrier.Retry(func() error {
			attempts++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("success after retry with retriable error", func(t *testing.T) {
		// Using a mock classifier that returns Retriable for specific errors
		mockClassifier := newMockClassifier(t)
		mockClassifier.EXPECT().ClassifyError(errors.New("temporary error")).Return(Retriable)
		mockClassifier.EXPECT().ClassifyError(nil).Return(NonRetriable)

		retrier := &pgRetrier{
			retryTimeouts: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
			classifier:    mockClassifier,
		}

		attempts := 0
		err := retrier.Retry(func() error {
			attempts++
			// Return a retriable error on first attempt, success on second
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
		mockClassifier.AssertExpectations(t)
	})

	t.Run("failure after all retries exhausted", func(t *testing.T) {
		retrier := NewPgRetrier([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond})

		expectedErr := errors.New("permanent error")
		err := retrier.Retry(func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})

	t.Run("retry with retriable error", func(t *testing.T) {
		// Using a mock classifier that returns Retriable for specific errors
		mockClassifier := newMockClassifier(t)
		mockClassifier.EXPECT().ClassifyError(errors.New("temporary error")).Return(Retriable)
		mockClassifier.EXPECT().ClassifyError(nil).Return(NonRetriable)

		retrier := &pgRetrier{
			retryTimeouts: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond},
			classifier:    mockClassifier,
		}

		attempts := 0
		err := retrier.Retry(func() error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
		mockClassifier.AssertExpectations(t)
	})

	t.Run("non-retriable error stops retries", func(t *testing.T) {
		// Using a mock classifier that returns NonRetriable for specific errors
		mockClassifier := newMockClassifier(t)
		mockClassifier.EXPECT().ClassifyError(errors.New("permanent error")).Return(NonRetriable)

		retrier := &pgRetrier{
			retryTimeouts: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
			classifier:    mockClassifier,
		}

		attempts := 0
		expectedErr := errors.New("permanent error")

		err := retrier.Retry(func() error {
			attempts++
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, attempts)
		mockClassifier.AssertExpectations(t)
	})

	t.Run("nil error handled correctly", func(t *testing.T) {
		retrier := NewPgRetrier([]time.Duration{10 * time.Millisecond})

		err := retrier.Retry(func() error {
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("no retry timeouts", func(t *testing.T) {
		retrier := NewPgRetrier([]time.Duration{})

		expectedErr := errors.New("error")
		err := retrier.Retry(func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})

	t.Run("empty retry timeouts slice", func(t *testing.T) {
		retrier := NewPgRetrier(nil)

		expectedErr := errors.New("error")
		err := retrier.Retry(func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})
}
