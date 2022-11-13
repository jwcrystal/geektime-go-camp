//go:build e2e

package web

import (
	"fmt"
	"testing"
)

func TestHttpServer(t *testing.T) {
	s := NewHttpServer()

	s.Get("/", func(ctx *Context) {
		ctx.Res.Write([]byte("hello, world"))
	})
	s.Get("/abc/*", func(ctx *Context) {
		ctx.Res.Write([]byte(fmt.Sprintf("hello, %s", ctx.Req.URL.Path)))
	})
	s.Get("/user/:username", func(ctx *Context) {
		ctx.Res.Write([]byte(fmt.Sprintf("hello, %s", ctx.PathParams["username"])))
	})
	err := s.Start(":8081")
	if err != nil {
		return
	}
}
