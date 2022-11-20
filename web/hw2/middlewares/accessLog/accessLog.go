package accessLog

import (
	"encoding/json"
	"fmt"
	"geektime-go/web"
)

type accessLog struct {
	Host       string `json:"host,omitempty"`
	Route      string `json:"route,omitempty"`
	HttpMethod string `json:"http_method,omitempty"`
	Path       string `json:"path,omitempty"`
}
type MiddlewareBuilder struct {
	logFunc func(log string)
}

func (m *MiddlewareBuilder) LogFunc(logFunc func(log string)) *MiddlewareBuilder {
	m.logFunc = logFunc
	return m
}

func NewBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(log string) {
			fmt.Println(log)
		},
	}
}

func (m *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			// log middleware
			defer func() {
				l := accessLog{
					Host:       ctx.Req.Host,
					Route:      ctx.MatchedRoute, // get the route after next() (route)
					HttpMethod: ctx.Req.Method,
					Path:       ctx.Req.URL.Path,
				}
				val, _ := json.Marshal(l)
				m.logFunc(string(val))
			}()
			next(ctx)
		}
	}
}
