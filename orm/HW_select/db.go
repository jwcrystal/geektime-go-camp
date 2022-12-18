package orm

import (
	"database/sql"
	"geektime-go/orm/HW_select/model"
)

type DB struct {
	model *model.Model
	db    *sql.DB
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
		db: db,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret, nil
}