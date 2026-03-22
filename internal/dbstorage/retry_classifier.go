package dbstorage

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mikeziminio/go-custom-metrics/internal/retrier"
)

// retryClassifier implements the retryClassifier interface for database-specific error handling.
//
// It classifies PostgreSQL errors into retriable and non-retriable categories:
//   - Retriable: connection issues (08xxx), transaction rollbacks (40xxx), deadlock (40P01)
//   - Non-retriable: all other errors including data errors (22xxx, 23xxx)
type retryClassifier struct{}

// newRetryClassifier creates a new instance of RetryClassifier.
//
// Returns a *RetryClassifier ready for use in classifying database errors.
func newRetryClassifier() *retryClassifier {
	return &retryClassifier{}
}

// ClassifyError determines whether a PostgreSQL error should be retried.
//
// Parameters:
//   - err: The error to classify
//
// Returns:
//   - retrier.Retriable: for connection failures, transaction rollbacks, deadlocks
//   - retrier.NonRetriable: for all other errors including data errors
func (*retryClassifier) ClassifyError(err error) retrier.ErrorClass {
	if err == nil {
		return retrier.NonRetriable
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return retrier.NonRetriable
	}

	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException, // 08000
		pgerrcode.ConnectionDoesNotExist, // 08003
		pgerrcode.ConnectionFailure:      // 08006
		return retrier.Retriable
	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return retrier.Retriable
	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return retrier.Retriable
	}

	return retrier.NonRetriable
}
