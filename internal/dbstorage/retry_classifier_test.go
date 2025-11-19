package dbstorage

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	"github.com/mikeziminio/go-custom-metrics/internal/retrier"
)

func TestRetryClassifier(t *testing.T) {
	classifier := newRetryClassifier()

	testCases := []struct {
		name          string
		error         error
		expectedClass retrier.ErrorClass
	}{
		{
			name:          "nil error",
			error:         nil,
			expectedClass: retrier.NonRetriable,
		},
		{
			name:          "non-PostgreSQL error",
			error:         errors.New("some regular error"),
			expectedClass: retrier.NonRetriable,
		},
		{
			name: "connection exception",
			error: &pgconn.PgError{
				Code: pgerrcode.ConnectionException,
			},
			expectedClass: retrier.Retriable,
		},
		{
			name: "connection does not exist",
			error: &pgconn.PgError{
				Code: pgerrcode.ConnectionDoesNotExist,
			},
			expectedClass: retrier.Retriable,
		},
		{
			name: "connection failure",
			error: &pgconn.PgError{
				Code: pgerrcode.ConnectionFailure,
			},
			expectedClass: retrier.Retriable,
		},
		{
			name: "transaction rollback",
			error: &pgconn.PgError{
				Code: pgerrcode.TransactionRollback,
			},
			expectedClass: retrier.Retriable,
		},
		{
			name: "serialization failure",
			error: &pgconn.PgError{
				Code: pgerrcode.SerializationFailure,
			},
			expectedClass: retrier.Retriable,
		},
		{
			name: "deadlock detected",
			error: &pgconn.PgError{
				Code: pgerrcode.DeadlockDetected,
			},
			expectedClass: retrier.Retriable,
		},
		{
			name: "cannot connect now",
			error: &pgconn.PgError{
				Code: pgerrcode.CannotConnectNow,
			},
			expectedClass: retrier.Retriable,
		},
		{
			name: "invalid catalog name (non-retriable)",
			error: &pgconn.PgError{
				Code: pgerrcode.InvalidCatalogName,
			},
			expectedClass: retrier.NonRetriable,
		},
		{
			name: "invalid schema name (non-retriable)",
			error: &pgconn.PgError{
				Code: pgerrcode.InvalidSchemaName,
			},
			expectedClass: retrier.NonRetriable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.ClassifyError(tc.error)
			assert.Equal(t, tc.expectedClass, result)
		})
	}
}
