package web

import "net/http"

type Context struct {
	Req *http.Request
	Res http.ResponseWriter
}
