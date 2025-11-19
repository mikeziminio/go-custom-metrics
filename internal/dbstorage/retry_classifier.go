package dbstorage

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mikeziminio/go-custom-metrics/internal/retrier"
)

type retryClassifier struct{}

func newRetryClassifier() *retryClassifier {
	return &retryClassifier{}
}

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
