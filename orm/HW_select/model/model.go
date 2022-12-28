package model

import "reflect"

type field struct {
	ColName string
	GoName  string
	Type    reflect.Type
}

type Model struct {
	TableName string
	FieldMap  map[string]*field
}

const (
	tagKeyColumn = "column"
)

// TableName 用戶實現這個接口得到自定義的表名
type TableName interface {
	TableName() string
}

//func ( *registry) parseModel(entity any) (*Model, error) {
//	typ := reflect.TypeOf(entity)
//	// 只支持一級指針
//	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
//		return nil, errs.ErrPointerOnly
//	}
//	typ = typ.Elem()
//	numField := typ.NumField()
//	fields := make(map[string]*field, numField)
//	for i := 0; i < numField; i++ {
//		typField := typ.Field(i)
//		// 直接用了 類型名稱 和 字段名稱 ( camel 命名)， 但 DB內通常採用 underline 命名
//		//fields[typField.Name] = &field{colName: typField.Name}
//		fields[typField.Name] = &field{ColName: underlineName(typField.Name)}
//	}
//	return &Model{
//		TableName: underlineName(typ.Name()),
//		FieldMap:  fields,
//	}, nil
//}
//
//func underlineName(field string) string {
//	var buf []byte
//	for i, c := range field {
//		if unicode.IsUpper(c) {
//			if i != 0 {
//				// e.g. testName -> test_name
//				buf = append(buf, '_')
//			}
//			buf = append(buf, byte(unicode.ToLower(c)))
//		} else {
//			buf = append(buf, byte(c))
//		}
//	}
//	return string(buf)
//}
