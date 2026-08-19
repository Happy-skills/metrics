package repository

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

func classify(err error) PGErrorClassification {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return classifyPgError(pgErr)
	}

	return NonRetriable
}

func classifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException, // 080000
		pgerrcode.ConnectionDoesNotExist, // 08003
		pgerrcode.ConnectionFailure,      // 080006
		pgerrcode.SerializationFailure,   // 400001
		pgerrcode.DeadlockDetected,       // 40P01
		pgerrcode.AdminShutdown,          // 57P01
		pgerrcode.TooManyConnections:     // 53300

		return Retriable
	}

	return NonRetriable
}
