package dbstorage

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type ErrorClass string

var (
	Retriable    ErrorClass = "retriable"
	NonRetriable ErrorClass = "non_retriable"
)

type classifier interface {
	ClassifyError(err error) ErrorClass
}

type pgClassifier struct{}

func NewPgClassifier() *pgClassifier {
	return &pgClassifier{}
}

func (c *pgClassifier) ClassifyError(err error) ErrorClass {
	if err == nil {
		return NonRetriable
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return NonRetriable
	}

	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException, // 08000
		pgerrcode.ConnectionDoesNotExist, // 08003
		pgerrcode.ConnectionFailure:      // 08006
		return Retriable
	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable
	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	return NonRetriable
}
