package dbstorage

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestPgClassifier(t *testing.T) {
	classifier := NewPgClassifier()

	testCases := []struct {
		name          string
		error         error
		expectedClass ErrorClass
	}{
		{
			name:          "nil error",
			error:         nil,
			expectedClass: NonRetriable,
		},
		{
			name:          "non-PostgreSQL error",
			error:         errors.New("some regular error"),
			expectedClass: NonRetriable,
		},
		{
			name: "connection exception",
			error: &pgconn.PgError{
				Code: pgerrcode.ConnectionException,
			},
			expectedClass: Retriable,
		},
		{
			name: "connection does not exist",
			error: &pgconn.PgError{
				Code: pgerrcode.ConnectionDoesNotExist,
			},
			expectedClass: Retriable,
		},
		{
			name: "connection failure",
			error: &pgconn.PgError{
				Code: pgerrcode.ConnectionFailure,
			},
			expectedClass: Retriable,
		},
		{
			name: "transaction rollback",
			error: &pgconn.PgError{
				Code: pgerrcode.TransactionRollback,
			},
			expectedClass: Retriable,
		},
		{
			name: "serialization failure",
			error: &pgconn.PgError{
				Code: pgerrcode.SerializationFailure,
			},
			expectedClass: Retriable,
		},
		{
			name: "deadlock detected",
			error: &pgconn.PgError{
				Code: pgerrcode.DeadlockDetected,
			},
			expectedClass: Retriable,
		},
		{
			name: "cannot connect now",
			error: &pgconn.PgError{
				Code: pgerrcode.CannotConnectNow,
			},
			expectedClass: Retriable,
		},
		{
			name: "invalid catalog name (non-retriable)",
			error: &pgconn.PgError{
				Code: pgerrcode.InvalidCatalogName,
			},
			expectedClass: NonRetriable,
		},
		{
			name: "invalid schema name (non-retriable)",
			error: &pgconn.PgError{
				Code: pgerrcode.InvalidSchemaName,
			},
			expectedClass: NonRetriable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.ClassifyError(tc.error)
			assert.Equal(t, tc.expectedClass, result)
		})
	}
}
