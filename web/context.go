package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req *http.Request
	// Resp 如果用戶直接使用這個
	// 那麼代表用戶繞開了 RespData 和 RespStatusCode
	// 部分 middleware 可能會無法運作
	Resp         http.ResponseWriter
	PathParams   map[string]string
	MatchedRoute string

	// 為了 middleware 讀寫用的
	RespStatusCode int
	RespData       []byte

	queryValue url.Values
	tplEngine  TemplateEngine
}

func (c *Context) Redirect(url string) {
	http.Redirect(c.Resp, c.Req, url, http.StatusFound)
}

func (c *Context) RespString(code int, msg string) error {
	c.RespStatusCode = code
	c.RespData = []byte(msg)
	return nil
}

func (c *Context) RespOk(msg string) error {
	c.RespStatusCode = http.StatusOK
	c.RespData = []byte(msg)
	return nil
}

func (c *Context) RespServerError(msg string) error {
	c.RespData = []byte(msg)
	c.RespStatusCode = http.StatusInternalServerError
	return nil
}

func (c *Context) Render(tplName string, data any) error {
	// 不要這樣做
	// tplName = tplName + ".gohtml"
	// tplName = tplName + c.tplPrefix
	var err error
	c.RespData, err = c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	c.RespStatusCode = http.StatusOK
	return nil
}

func (c *Context) SetCookie(ck *http.Cookie) {
	// 不推薦
	// ck.SameSite = c.cookieSameSite
	http.SetCookie(c.Resp, ck)
}

func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

func (c *Context) RespJSON(code int, val any) error {
	dataJSON, err := json.Marshal(val)
	if err != nil {
		return err
	}
	// c.Resp.Header().Set("Content-Type", "application/json")
	// c.Resp.Header().Set("Content-Length", strconv.Itoa(len(data)))
	c.RespStatusCode = code
	c.RespData = dataJSON
	return nil
}

// 解決大多數人的需求
func (c *Context) BindJSON(val any) error {
	if c.Req.Body == nil {
		return errors.New("web: body為空")
	}
	// bs, _:= io.ReadAll(c.Req.Body)
	// json.Unmarshal(bs, val)
	decoder := json.NewDecoder(c.Req.Body)
	// useNumber => 数字就是用 Number 来表示
	// 否则默认是 float64
	// if jsonUseNumber {
	// 	decoder.UseNumber()
	// }

	// 如果有多的字段(未知字段)，就會出錯
	// e.g. 原有的 json struct 只有 Name 和 Email
	// 得到的 JSON 多了一個 Age 字段，就會報錯
	// decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

// FormValue(key1)
// FormValue(key2)
func (c *Context) FormValue(key string) (string, error) {
	err := c.Req.ParseForm()
	if err != nil {
		return "", err
	}
	return c.Req.FormValue(key), nil
}

func (c *Context) QueryValue(key string) (string, error) {
	if c.queryValue == nil {
		c.queryValue = c.Req.URL.Query()
	}
	vals, ok := c.queryValue[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return vals[0], nil

	// 如果採用 Get()， 用戶無法區分是否有值，還是空字串
	// return c.queryValues.Get(key), nil
}

func (c *Context) QueryValueV1(key string) StringValue {
	if c.queryValue == nil {
		c.queryValue = c.Req.URL.Query()
	}
	vals, ok := c.queryValue[key]
	if !ok {
		return StringValue{err: errors.New("web: key 不存在")}
	}
	return StringValue{
		val: vals[0],
	}
}

func (c *Context) PathValueV1(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{err: errors.New("web: key 不存在")}
	}
	return StringValue{val: val}
}
func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return val, nil
}

// StringValue 無法使用泛型，因為在創建時候我們不知道用戶需要說明作為 T
//
//	type StringValue[T any] struct {
//		val string
//		err error
//	}
type StringValue struct {
	val string
	err error
}

func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}

// func (s StringValue[T]) As() (T, error) {
// }

// func (s StringValue) AsInt32() (int64, error) {
//
// }

// func (s StringValue) AsInt() (int64, error) {
//
// }
