package orm

import (
	"context"
	"database/sql"
	"errors"
	"geektime-go/orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			// aggregate 聚合函數, 型態 2
			name: "avg_type 2",
			q: NewSelector[TestModel](db).Select(Avg("Age").As("avg_age")).
				GroupBy(C("FirstName")).
				Having(Raw("`avg_age` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT AVG(`age`) AS `avg_age` FROM `test_model` GROUP BY `first_name` HAVING `avg_age` < ?;",
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
	db := memoryDB(t)
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
			q:    NewSelector[TestModel](db).From("`test_table`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table`;",
				Args: nil,
			},
		},
		{
			name: "empty from",
			q:    NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name: "from db",
			q:    NewSelector[TestModel](db).From("`test_db`.`test_table`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_table`;",
				Args: nil,
			},
		},
		{
			name: "single and simple predicate",
			q: NewSelector[TestModel](db).From("`test_table`").
				Where(C("Id").Eq(1)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE `id` = ?;",
				Args: []any{1},
			},
		},
		{
			name: "multi-predicate",
			q: NewSelector[TestModel](db).From("`test_table`").
				Where(C("Id").Eq(1), C("Age").Eq(12)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE (`id` = ?) AND (`age` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "and",
			q: NewSelector[TestModel](db).From("`test_table`").
				Where(C("Id").Eq(1).And(C("Age").Eq(12))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE (`id` = ?) AND (`age` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "or",
			q: NewSelector[TestModel](db).From("`test_table`").
				Where(C("Id").Eq(1).Or(C("Id").Eq(12))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_table` WHERE (`id` = ?) OR (`id` = ?);",
				Args: []any{1, 12},
			},
		},
		{
			name: "not",
			q: NewSelector[TestModel](db).From("`test_table`").
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
			q: NewSelector[TestModel](db).
				Where(Raw("`age` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` < ?;",
				Args: []any{18},
			},
		},
		{
			// 非法列，需要擷取 DB 的 schema
			name:    "invalid column",
			q:       NewSelector[TestModel](db).Where(Not(C("Invalid").Eq(1))),
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

func TestSelector_Select(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantError error
	}{
		{
			name: "all",
			q:    NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:      "invalid column",
			q:         NewSelector[TestModel](db).Select(Avg("Invalid")),
			wantError: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "partial column",
			q:    NewSelector[TestModel](db).Select(C("Id")),
			wantQuery: &Query{
				SQL: "SELECT `id` FROM `test_model`;",
			},
		},
		{
			name: "avg",
			q:    NewSelector[TestModel](db).Select(Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name: "raw expression",
			q:    NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT `first_name`)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT `first_name`) FROM `test_model`;",
			},
		},
		{
			name: "alias",
			q: NewSelector[TestModel](db).
				Select(C("Id").As("my_id"),
					Avg("Age").As("avg_age")),
			wantQuery: &Query{
				SQL: "SELECT `id` AS `my_id`,AVG(`age`) AS `avg_age` FROM `test_model`;",
			},
		},
		// WHERE 忽略別名
		{
			name: "where ignore alias",
			q: NewSelector[TestModel](db).
				Where(C("Id").As("my_id").Lt(100)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` < ?;",
				Args: []any{100},
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

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	//if err != nil {
	//	t.Fatal(err)
	//}
	defer func() {
		_ = mockDB.Close()
	}()
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	//if err != nil {
	//	t.Fatal(err)
	//}

	//// query error
	//mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))
	//
	//// no rows
	//rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	//mock.ExpectQuery("SELECT .*").WillReturnRows(rows)
	//
	//// get data
	//rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	//rows.AddRow("1", "Tom", "18", "Jerry")
	//mock.ExpectQuery("SELECT .*").WillReturnRows(rows)
	//
	//// scan error
	//rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	//rows.AddRow("abc", "Tom", "18", "Jerry")
	//mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	testCases := []struct {
		name      string
		query     string
		mockError error
		mockRows  *sqlmock.Rows
		wantValue *TestModel
		wantError error
	}{
		{
			name:      "query error",
			mockError: errors.New("invalid query"),
			wantError: errors.New("invalid query"),
			// 正則表達式
			query: "SELECT .*",
		},
		{
			name:      "no rows",
			wantError: errs.ErrNoRows,
			// 正則表達式
			query:    "SELECT .*",
			mockRows: sqlmock.NewRows([]string{"id"}),
		},
		{
			name:      "too many column",
			wantError: errs.ErrTooManyReturnedColumns,
			query:     "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				ret := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "extra_column"})
				ret.AddRow([]byte("1"), []byte("Wang"), []byte("18"), []byte("Jared"), []byte("nothing"))
				return ret
			}(),
		},
		{
			name:  "get data",
			query: "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				res.AddRow([]byte("1"), []byte("Wang"), []byte("18"), []byte("Jared"))
				return res
			}(),
			wantValue: &TestModel{
				Id:        1,
				FirstName: "Wang",
				Age:       18,
				LastName:  &sql.NullString{String: "Jared", Valid: true},
			},
		},
	}

	// mock 查詢和實際查詢需要完全一致
	// 跟上面的 mockSQL 是相同行為
	for _, tc := range testCases {
		expr := mock.ExpectQuery(tc.query)
		if tc.mockError != nil {
			expr.WillReturnError(tc.mockError)
		} else {
			expr.WillReturnRows(tc.mockRows)
		}
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//var ret any
			ret, err := NewSelector[TestModel](db).Get(context.Background())
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantValue, ret)
		})
	}
}
