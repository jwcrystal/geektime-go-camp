//go:build e2e

package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHttpServer(t *testing.T) {
	s := NewHttpServer()

	s.UseV1(http.MethodGet, "/", func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			fmt.Println("hello, world")
			ctx.Resp.Write([]byte(fmt.Sprintf("hello 111 %s\n", ctx.Req.URL.Path)))
			next(ctx)
		}
	})
	s.Get("/", func(ctx *Context) {
		fmt.Println("hello, world")
		ctx.Resp.Write([]byte("hello, world"))
	})
	s.Get("/abc/*", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %s", ctx.Req.URL.Path)))
	})
	s.Get("/user/:username", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %s", ctx.PathParams["username"])))
	})

	err := s.Start(":8081")
	if err != nil {
		return
	}
}
