package retrier

import (
	"fmt"
	"time"
)

func ExampleRetrier_Retry_successFirstAttempt() {
	classifier := NewDefaultRetryClassifier()
	retrier := NewRetrier([]time.Duration{100 * time.Millisecond}, classifier)

	attempts := 0
	op := func() error {
		attempts++
		return nil
	}

	err := retrier.Retry(op)
	fmt.Printf("Operation succeeded after %d attempts, error: %v\n", attempts, err)
	// Output:
	// Operation succeeded after 1 attempts, error: <nil>
}

func ExampleRetrier_Retry_onError() {
	classifier := NewDefaultRetryClassifier()
	retrier := NewRetrier([]time.Duration{10 * time.Millisecond}, classifier)

	attempts := 0
	op := func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("some error")
		}
		return nil
	}

	err := retrier.Retry(op)
	fmt.Printf("Operation succeeded after %d attempts, error: %v\n", attempts, err)
	// Output:
	// Operation succeeded after 3 attempts, error: <nil>
}
