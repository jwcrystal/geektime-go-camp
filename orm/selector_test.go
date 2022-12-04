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

func TestSelector_Build(t *testing.T) {
	var testCases = []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "no from",
			q:    &Selector[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name: "with from",
			q:    (&Selector[TestModel]{}).From("`test_table`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table`;",
				Args: nil,
			},
		},
		{
			name: "empty from",
			q:    (&Selector[TestModel]{}).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name: "from db",
			q:    (&Selector[TestModel]{}).From("`test_db`.`test_table`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_table`;",
				Args: nil,
			},
		},
		{
			name: "single and simple predicate",
			q: (&Selector[TestModel]{}).From("`test_table`").
				Where(C("Id").Eq(1)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE `id` = ?;",
				Args: []any{1},
			},
		},
		{
			name: "multi-predicate",
			q: (&Selector[TestModel]{}).From("`test_table`").
				Where(C("Id").Eq(1), C("Age").Eq(12)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE (`id` = ?) AND (`age` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "and",
			q: (&Selector[TestModel]{}).From("`test_table`").
				Where(C("Id").Eq(1).And(C("Age").Eq(12))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE (`id` = ?) AND (`age` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "or",
			q: (&Selector[TestModel]{}).From("`test_table`").
				Where(C("Id").Eq(1).Or(C("Id").Eq(12))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE (`id` = ?) OR (`id` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "not",
			q: (&Selector[TestModel]{}).From("`test_table`").
				Where(Not(C("Id").Eq(1))),
			wantQuery: &Query{
				// NOT 前面有兩個空格，因為沒有特別處理
				SQL:  "SELECT * FROM `test_table` WHERE  NOT (`id` = ?);",
				Args: []any{1},
			},
		},
		{
			// 使用 RawExpr
			name: "raw expression",
			q: (&Selector[TestModel]{}).
				Where(Raw("`age` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` < ?);",
				Args: []any{18},
			},
		},
		{
			// 非法列，需要擷取 DB 的 schema
			name:    "invalid column",
			q:       (&Selector[TestModel]{}).Where(Not(C("Invalid").Eq(1))),
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
