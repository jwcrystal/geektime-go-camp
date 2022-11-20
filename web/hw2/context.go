package web

import "net/http"

type Context struct {
	Req          *http.Request
	Res          http.ResponseWriter
	PathParams   map[string]string
	MatchedRoute string

	ResStatusCode int
	ResData       []byte
}
