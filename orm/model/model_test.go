package model

import (
	"database/sql"
	"geektime-go/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func Test_parseModel(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantErr   error
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
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
			name:    "map",
			entity:  map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "basic types",
			entity:  0,
			wantErr: errs.ErrPointerOnly,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model, err := ParseModel(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, model)
		})
	}
}
