package valuer

import (
	"database/sql"
	"geektime-go/orm/model"
	"reflect"
)

// reflectValue 基於反射的 Value
type reflectValue struct {
	val  reflect.Value
	meta *model.Model
}

// NewReflectValue 返回一個封裝且基於反射的 Value 對象
func NewReflectValue(val any, meta *model.Model) Value {
	return &reflectValue{
		val:  reflect.ValueOf(val).Elem(),
		meta: meta,
	}
}

func (r *reflectValue) SetColumns(rows *sql.Rows) error {
	//tp := new(T)
	//cs, err := rows.Columns()
	//if err != nil {
	//	return err
	//}
	//vals := make([]any, 0, len(cs))
	//// 針對 column 產生 model 字段的類型指針
	//for _, c := range cs {
	//	for _, fd := range s.model.FieldMap {
	//		if fd.ColName == c {
	//			// 反射建立一個實例
	//			// 這個實例是原本類型的指針類型
	//			// e.g. fd.Type = int, val 則是 *int
	//			val := reflect.New(fd.Type)
	//			vals = append(vals, val.Interface())
	//		}
	//	}
	//}
	//// 第一個考量: 類型要匹配
	//// 第二個考量: 順序要匹配
	//// e.g.
	//// SELECT id, first_name, age, last_name
	//// SELECT first_name, id, age, last_name
	//rows.Scan(vals...)
	//// 把 vals 塞回去 結果 tp 裡面
	//tpValue := reflect.ValueOf(tp)
	//for i, c := range cs {
	//	for _, fd := range s.model.FieldMap {
	//		if fd.ColName == c {
	//			tpValue.Elem().FieldByName(fd.GoName).
	//				Set(reflect.ValueOf(vals[i]).Elem())
	//		}
	//	}
	//}
	return nil
}
