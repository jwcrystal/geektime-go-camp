package model

import (
	"geektime-go/orm/HW_select/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegistry_Register(t *testing.T) {
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
				FieldMap: map[string]*field{
					"Id": {
						ColName: "id",
					},
					"FirstName": {
						ColName: "first_name",
					},
					"Age": {
						ColName: "age",
					},
					"LastName": {
						ColName: "last_name",
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
				FieldMap: map[string]*field{
					"ID": {
						ColName: "id",
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
				FieldMap: map[string]*field{
					"FirstName": {
						ColName: "first_name",
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
				FieldMap: map[string]*field{
					"Name": {
						ColName: "name",
					},
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
				FieldMap: map[string]*field{
					"Name": {
						ColName: "name",
					},
				},
			},
		},
		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
				FieldMap: map[string]*field{
					"Name": {
						ColName: "name",
					},
				},
			},
		},
	}
	r := &registry{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model, err := r.Register(tc.entity)
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
