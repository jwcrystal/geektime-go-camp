package orm

import (
	"database/sql"
	"geektime-go/orm/internal/valuer"
	"geektime-go/orm/model"
)

type DB struct {
	r          model.Registry
	db         *sql.DB
	valCreator valuer.Creator
}

type DBOption func(*DB)

func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	ret := &DB{
		r:          model.NewRegistry(),
		db:         db,
		valCreator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret, nil
}

func DBWithRegitry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBUseReflectValuer() DBOption {
	return func(db *DB) {
		db.valCreator = valuer.NewReflectValue
	}
}
