package orm

import (
	"geektime-go/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_OffsetLimit(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantError error
	}{
		{
			name: "offset only",
			q:    NewSelector[TestModel](db).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` OFFSET ?;",
				Args: []any{10},
			},
		},
		{
			name: "limit only",
			q:    NewSelector[TestModel](db).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ?;",
				Args: []any{10},
			},
		},
		{
			name: "offset and limit",
			q:    NewSelector[TestModel](db).Offset(10).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` OFFSET ? LIMIT ?;",
				Args: []any{10, 10},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_OrderBy(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantError error
	}{
		{
			name: "column",
			q:    NewSelector[TestModel](db).OrderBy(Asc("Age")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` ORDER BY `age` ASC;",
			},
		},
		{
			name: "columns",
			q:    NewSelector[TestModel](db).OrderBy(Asc("Age"), Desc("Id")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` ORDER BY `age` ASC,`id` DESC;",
			},
		},
		{
			name:      "invalid column",
			q:         NewSelector[TestModel](db).OrderBy(Asc("Invalid")),
			wantError: errs.NewErrUnknownField("Invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_Having(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantError error
	}{
		{
			name: "none",
			q:    NewSelector[TestModel](db).GroupBy(C("Age")).Having(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			name: "single",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("LastName").Eq("Jared")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING `last_name` = ?;",
				Args: []any{"Jared"},
			},
		},
		{
			name: "multiple",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("LastName").Eq("Jared"), C("FirstName").Eq("Wang")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING (`last_name` = ?) AND (`first_name` = ?);",
				Args: []any{"Jared", "Wang"},
			},
		},
		{
			// aggregate 聚合函數，大同小異，以 AVG 為 testcase
			name: "avg",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(Avg("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING AVG(`age`) = ?;",
				Args: []any{18},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_GroundBy(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantError error
	}{
		{
			name: "none",
			q:    NewSelector[TestModel](db).GroupBy(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name: "single",
			q:    NewSelector[TestModel](db).GroupBy(C("Age")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			name: "multiple",
			q:    NewSelector[TestModel](db).GroupBy(C("Age"), C("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`,`first_name`;",
			},
		},
		{
			// 不存在
			name:      "invalid column",
			q:         NewSelector[TestModel](db).GroupBy(C("Invalid")),
			wantError: errs.NewErrUnknownField("Invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
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
