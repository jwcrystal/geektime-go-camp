package model

import (
	"database/sql"
	"geektime-go/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestModelWithTableName(t *testing.T) {
	testCases := []struct {
		name          string
		entity        any
		opt           Option
		Field         string
		wantTableName string
		wantError     error
	}{
		{
			name:          "empty string",
			entity:        &TestModel{},
			opt:           WithTableName(""),
			wantTableName: "",
		},
		{
			name:          "table name",
			entity:        &TestModel{},
			opt:           WithTableName("test_model_t"),
			wantTableName: "test_model_t",
		},
	}

	r := NewRegistry().(*registry)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model, err := r.Register(tc.entity, tc.opt)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantTableName, model.TableName)
		})
	}
}

func TestWithColumnName(t *testing.T) {
	testCases := []struct {
		name        string
		entity      any
		opt         Option
		field       string
		wantColName string
		wantError   error
	}{
		{
			name:        "new name",
			entity:      &TestModel{},
			opt:         WithColumnName("FirstName", "first_name_t"),
			field:       "FirstName",
			wantColName: "first_name_t",
		},
		{
			name:        "empty new name",
			entity:      &TestModel{},
			opt:         WithColumnName("FirstName", ""),
			field:       "FirstName",
			wantColName: "",
		},
		{
			name:      "invalid field name",
			entity:    &TestModel{},
			opt:       WithColumnName("FirstNameXXX", "first_name"),
			field:     "FirstNameXXX",
			wantError: errs.NewErrUnknownField("FirstNameXXX"),
		},
	}

	r := NewRegistry().(*registry)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model, err := r.Register(tc.entity, tc.opt)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			fd := model.FieldMap[tc.field]
			assert.Equal(t, tc.wantColName, fd.ColName)
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantError error
	}{
		{
			name:      "struct",
			entity:    TestModel{},
			wantError: errs.ErrPointerOnly,
		},
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
						Type:    reflect.TypeOf(int64(0)),
						GoName:  "Id",
					},
					"FirstName": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
					"Age": {
						ColName: "age",
						Type:    reflect.TypeOf(int8(0)),
						GoName:  "Age",
					},
					"LastName": {
						ColName: "last_name",
						Type:    reflect.TypeOf(&sql.NullString{}),
						GoName:  "LastName",
					},
				},
			},
		},
		{
			// 多级指针
			name: "multiple pointer",
			// 因為 Go 編譯器緣故，採用下方樣式呈現
			entity: func() any {
				val := &TestModel{}
				return &val
			}(),
			wantError: errs.ErrPointerOnly,
		},
		{
			name:      "map",
			entity:    map[string]string{},
			wantError: errs.ErrPointerOnly,
		},
		{
			name:      "slice",
			entity:    []string{},
			wantError: errs.ErrPointerOnly,
		},
		{
			name:      "basic type",
			entity:    0,
			wantError: errs.ErrPointerOnly,
		},
		// 標籤測試用例
		{
			name: "column tag",
			entity: func() any {
				// 把測試結構體定義在匿名方法內，防止被其他用例訪問
				type ColumnTag struct {
					ID uint64 `orm:"column=id"`
				}
				return &ColumnTag{}
			}(),
			wantModel: &Model{
				TableName: "column_tag",
				FieldMap: map[string]*Field{
					"ID": {
						ColName: "id",
						GoName:  "ID",
						Type:    reflect.TypeOf(uint64(0)),
					},
				},
			},
		},
		{
			// 用戶設置了 column，但沒有賦值
			name: "invalid tag",
			entity: func() any {
				type InvalidTag struct {
					FirstName uint64 `orm:"column"`
				}
				return &InvalidTag{}
			}(),
			wantError: errs.NewErrInvalidTagContent("column"),
		},
		{
			// 用戶設置其他內容 Tag，會自動過濾掉
			name: "ignore tag",
			entity: func() any {
				type IgnoreTag struct {
					FirstName uint64 `orm:"abc=abc"`
				}
				return &IgnoreTag{}
			}(),
			wantModel: &Model{
				TableName: "ignore_tag",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Type:    reflect.TypeOf(uint64(0)),
					},
				},
			},
		},
		// 利用接口自定義模型訊息
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
				FieldMap: map[string]*Field{
					"Name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
				FieldMap: map[string]*Field{
					"Name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
				FieldMap: map[string]*Field{
					"Name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
	}
	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, model)
		})
	}
}

type CustomTableName struct {
	Name string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	Name string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

type EmptyTableName struct {
	Name string
}

func (c *EmptyTableName) TableName() string {
	return ""
}
