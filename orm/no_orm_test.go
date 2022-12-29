package orm

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func (TestModel) CreateSQL() string {
	return `CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL)`
}

// memoryDB 返回一個基於內存的 ORM，它使用的是 sqlite3 內存模式
func memoryDB(t *testing.T) *DB {
	orm, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	if err != nil {
		t.Fatal(err)
	}
	return orm
}

func memoryDBWithDB(db string, t *testing.T) *DB {
	orm, err := Open("sqlite3", fmt.Sprintf("file:%s.db?cache=shared&mode=memory", db))
	if err != nil {
		t.Fatal(err)
	}
	return orm
}
