package valuer

import (
	"context"
	"database/sql"
	"geektime-go/orm/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReflectValue_SetColumns(t *testing.T) {
	testSetColumns(t, NewReflectValue)
}

func testSetColumns(t *testing.T, creator Creator) {
	testCases := []struct {
		name string
		//columns   map[string][]byte
		rows   *sqlmock.Rows
		entity any //*test.SimpleStruct
		//wantValue any //*test.SimpleStruct
		wantEntity any
		wantError  error
	}{
		{
			name:   "set columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow("1", "Wang", "18", "Jared")
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Wang",
				Age:       18,
				LastName: &sql.NullString{
					String: "Jared",
					Valid:  true,
				},
			},
		},
		{
			name:   "different order",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow("1", "Jared", "18", "Wang")
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Jared",
				Age:       18,
				LastName: &sql.NullString{
					String: "Wang",
					Valid:  true,
				},
			},
		},
		{
			name:   "partial columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "last_name"})
				rows.AddRow("1", "Jared")
				return rows
			}(),
			wantEntity: &TestModel{
				Id: 1,
				LastName: &sql.NullString{
					String: "Jared",
					Valid:  true,
				},
			},
		},
	}
	r := model.NewRegistry()

	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// 構造 rows
			mockRows := tc.rows
			//cols := make([]string, 0, len(tc.columns))
			//colVals := make([]driver.Value, 0, len(tc.columns))
			//for i, v := range tc.columns {
			//	cols = append(cols, i)
			//	colVals = append(colVals, v)
			//}
			mock.ExpectQuery("SELECT .*").
				WillReturnRows(mockRows)
			rows, err := mockDB.QueryContext(context.Background(), "SELECT .*")
			require.NoError(t, err)

			if !rows.Next() {
				return
			}

			m, err := r.Get(&TestModel{})
			//m, err := r.Get(&test.SimpleStruct{})
			require.NoError(t, err)
			val := creator(tc.entity, m)
			err = val.SetColumns(rows)

			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			if tc.wantError != nil {
				t.Fatalf("期望得到錯誤，但是並沒有發生 %v", tc.wantError)
			}
			assert.Equal(t, tc.wantEntity, tc.entity)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
