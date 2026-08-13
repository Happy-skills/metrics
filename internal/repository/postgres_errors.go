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
	// Class 08 - Connection errors - 08xxxx
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:

		return Retriable
	}

	return NonRetriable
}
