package valuer

import (
	"database/sql"
	"geektime-go/orm/model"
)

// Value 是對結構體實例的內部抽象
// 就是要把返回的結構體，包裝成一個 Value 對象
type Value interface {
	// SetColumns 設定新值
	SetColumns(rows *sql.Rows) error
}

type Creator func(val any, meta *model.Model) Value

//// ResultSetHandler 另一種可行的設計方案
//type ResultSetHandler interface {
//	// SetColumns 設定新值，column 是列名
//	SetColumns(val any, rows *sql.Rows) error
//}
