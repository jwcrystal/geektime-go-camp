package orm

import (
	"database/sql"
	"geektime-go/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func TestDelete_Build(t *testing.T) {
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "no where",
			q:    (&Deleter[TestModel]{}),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name: "where",
			q:    (&Deleter[TestModel]{}).Where(C("Id").Eq(1)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{1},
			},
		},
		{
			name: "from",
			q:    (&Deleter[TestModel]{}).From("`test_model`").Where(C("Id").Eq(16)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{16},
			},
		},
		{
			name: "and",
			q: (&Deleter[TestModel]{}).From("`test_table`").
				Where(C("Id").Eq(1).And(C("Age").Eq(12))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_table` WHERE (`id` = ?) AND (`age` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "or",
			q: (&Deleter[TestModel]{}).From("`test_table`").
				Where(C("Id").Eq(1).Or(C("Id").Eq(12))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_table` WHERE (`id` = ?) OR (`id` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "not",
			q: (&Deleter[TestModel]{}).From("`test_table`").
				Where(Not(C("Id").Eq(1))),
			wantQuery: &Query{
				// NOT 前面有兩個空格，因為沒有特別處理
				SQL:  "DELETE FROM `test_table` WHERE  NOT (`id` = ?);",
				Args: []any{1},
			},
		},
		{
			// 非法列，需要擷取 DB 的 schema
			name:    "invalid column",
			q:       (&Deleter[TestModel]{}).Where(Not(C("Invalid").Eq(1))),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}
