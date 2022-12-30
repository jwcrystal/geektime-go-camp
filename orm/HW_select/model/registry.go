package model

import (
	"geektime-go/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type Option func(m *Model) error

// Registry 元數據註冊中心的抽象
type Registry interface {
	// Get 查找元數據
	Get(val any) (*Model, error)
	// Register 註冊模型
	Register(val any, opts ...Option) (*Model, error)
}

// registry 代表是元數據的註冊中心, 基於標籤和接口的實現
// 因目前只有一個實現，可以暫時維持 private
type registry struct {
	// 普通 map，併發讀寫場景會 crash，需要加入鎖 lock 機制
	// lock sync.RWMutex
	// models map[reflect.Type]*model

	// 同時解析時候，會出現覆蓋問題
	// 在此假設解析的元數據是不變的，所以問題不大
	models sync.Map
}

func NewRegistry() Registry {
	return &registry{}
}

// Get 會先查找 models。沒有找到就會開始解析 model，解析完放回去 models
func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	model, ok := r.models.Load(typ)
	if ok {
		return model.(*Model), nil
	}
	return r.Register(val)
}

// func (r *registry) Get1(val any) (*Model, error) {
// 	typ := reflect.TypeOf(val)
// 	r.lock.RLock()
// 	m, ok := r.models[typ]
// 	r.lock.RUnlock()
// 	if ok {
// 		return m, nil
// 	}
//
// 	r.lock.Lock()
// 	defer r.lock.Unlock()
// 	m, ok = r.models[typ]
// 	if ok {
// 		return m, nil
// 	}
//  m, err := r.parseModel(val)
//	//m, err := r.Register(val)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.models[typ] = m
// 	return m, nil
// }

func (r *registry) Register(entity any, opts ...Option) (*Model, error) {
	model, err := r.parseModel(entity)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err = opt(model)
		if err != nil {
			return nil, err
		}
	}
	typ := reflect.TypeOf(entity)
	//r.models[typ] = model => Use Map[]
	r.models.Store(typ, model)
	return model, nil
}

// parseModel 支持從標籤種提取自定義設置
// 標籤形式 orm:"key1=value1,key2=value2"
func (r *registry) parseModel(entity any) (*Model, error) {
	typ := reflect.TypeOf(entity)
	// 只支持一級指針
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numFields := typ.NumField()
	fields := make(map[string]*field, numFields)
	//colMap := make(map[string]*field, numField)
	for i := 0; i < numFields; i++ {
		typField := typ.Field(i)
		// 直接用了 類型名稱 和 字段名稱 ( camel 命名)， 但 DB 內通常採用 underline 命名
		//fields[typField.Name] = &field{colName: typField.Name}
		//fields[typField.Name] = &field{ColName: underlineName(typField.Name)}
		// 取得結構體的 Tag
		tags, err := r.parseTag(typField.Tag)
		if err != nil {
			return nil, err
		}
		colName := tags[tagKeyColumn]
		if colName == "" {
			colName = underlineName(typField.Name)
		}
		fields[typField.Name] = &field{ColName: colName}
		//f := &field{
		//	colName,
		//}
	}
	var tableName string
	if tn, ok := entity.(TableName); ok {
		tableName = tn.TableName()
	}
	if tableName == "" {
		tableName = underlineName(typ.Name())
	}

	return &Model{
		TableName: tableName,
		FieldMap:  fields,
	}, nil
}

// parseTag 此處的 map 就是我們希望的 tag 標籤 "column： xxx" 型態
func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		return map[string]string{}, nil
	}
	// 初始化容量為我們支持的 key 數量
	ret := make(map[string]string, 1)

	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		ret[kv[0]] = kv[1]
	}
	return ret, nil
}

func underlineName(field string) string {
	var buf []byte
	for i, c := range field {
		if unicode.IsUpper(c) {
			if i != 0 {
				// e.g. testName -> test_name
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(c)))
		} else {
			buf = append(buf, byte(c))
		}
	}
	return string(buf)
}
