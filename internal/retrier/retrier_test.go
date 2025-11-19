package retrier

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrier(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		retrier := NewRetrier([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond}, NewDefaultRetryClassifier())

		attempts := 0
		err := retrier.Retry(func() error {
			attempts++
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("success after retry with retriable error", func(t *testing.T) {
		mockClassifier := NewMockRetryClassifier(t)
		mockClassifier.EXPECT().ClassifyError(errors.New("temporary error")).Return(Retriable)
		mockClassifier.EXPECT().ClassifyError(nil).Return(NonRetriable)

		retrier := NewRetrier([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond}, mockClassifier)

		attempts := 0
		err := retrier.Retry(func() error {
			attempts++
			// Возвращаем повторяемую ошибку на первой попытке
			// и нет ошибки на второй попытке
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("failure after all retries exhausted", func(t *testing.T) {
		retrier := NewRetrier([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond}, NewDefaultRetryClassifier())

		expectedErr := errors.New("permanent error")
		err := retrier.Retry(func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})

	t.Run("retry with retriable error", func(t *testing.T) {
		mockClassifier := NewMockRetryClassifier(t)
		mockClassifier.EXPECT().ClassifyError(errors.New("temporary error")).Return(Retriable)
		mockClassifier.EXPECT().ClassifyError(nil).Return(NonRetriable)

		retrier := NewRetrier([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}, mockClassifier)

		attempts := 0
		err := retrier.Retry(func() error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("non-retriable error stops retries", func(t *testing.T) {
		mockClassifier := NewMockRetryClassifier(t)
		mockClassifier.EXPECT().ClassifyError(errors.New("permanent error")).Return(NonRetriable)

		retrier := NewRetrier([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond}, mockClassifier)

		attempts := 0
		expectedErr := errors.New("permanent error")

		err := retrier.Retry(func() error {
			attempts++
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("nil error handled correctly", func(t *testing.T) {
		retrier := NewRetrier([]time.Duration{10 * time.Millisecond}, NewDefaultRetryClassifier())

		err := retrier.Retry(func() error {
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("no retry timeouts", func(t *testing.T) {
		retrier := NewRetrier([]time.Duration{}, NewDefaultRetryClassifier())

		expectedErr := errors.New("error")
		err := retrier.Retry(func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})

	t.Run("empty retry timeouts slice", func(t *testing.T) {
		retrier := NewRetrier(nil, NewDefaultRetryClassifier())

		expectedErr := errors.New("error")
		err := retrier.Retry(func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})
}
