package dbstorage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DBStorage struct {
	connString string
}

func New(connString string) *DBStorage {
	return &DBStorage{
		connString: connString,
	}
}

func (s *DBStorage) Ping(ctx context.Context) error {
	_, err := pgx.Connect(ctx, s.connString)
	return err
}
