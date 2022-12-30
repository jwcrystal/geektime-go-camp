package valuer

import (
	"database/sql/driver"
	"geektime-go/orm/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"testing"
)

func BenchmarkSetColumns(b *testing.B) {

	fn := func(b *testing.B, creator Creator) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(b, err)
		defer func() { _ = mockDB.Close() }()

		// 需要跑 N 次，所以需要 N 行
		mockRows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
		row := []driver.Value{"1", "Wang", "18", "Jared"}
		for i := 0; i < b.N; i++ {
			mockRows.AddRow(row...)
		}
		mock.ExpectQuery("SELECT .*").WillReturnRows(mockRows)
		rows, err := mockDB.Query("SELECT .*")
		require.NoError(b, err)

		r := model.NewRegistry()
		m, err := r.Get(&TestModel{})
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows.Next()
			val := creator(&TestModel{}, m)
			_ = val.SetColumns(rows)
		}
	}

	b.Run("reflect", func(b *testing.B) {
		fn(b, NewReflectValue)
	})
	b.Run("unsafe", func(b *testing.B) {
		fn(b, NewUnsafeValue)
	})
}
