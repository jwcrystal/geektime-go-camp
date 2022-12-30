package valuer

import (
	"database/sql"
	"geektime-go/orm/internal/errs"
	"geektime-go/orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	addr  unsafe.Pointer
	model *model.Model
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(val any, model *model.Model) Value {
	return unsafeValue{
		addr:  unsafe.Pointer(reflect.ValueOf(val).Pointer()),
		model: model,
	}
}

func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	// 起始地址
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	vals := make([]any, 0, len(cs))
	for _, c := range cs {
		fd, ok := u.model.ColumnMap[c]
		if !ok {
			return errs.ErrTooManyReturnedColumns
		}

		// 計算字段地址
		// 計算字段偏移量
		// 得到字段真實地址：字段起始位址 + 字段偏移量
		// 在真實地址創建對象
		fdAddress := unsafe.Pointer(uintptr(u.addr) + fd.Offset)
		// 反射建立一個實例
		// 這個實例是原本類型的指針類型
		// e.g. fd.Type = int, val 則是 *int
		val := reflect.NewAt(fd.Type, fdAddress)
		vals = append(vals, val.Interface())
	}

	return rows.Scan(vals...)
}
