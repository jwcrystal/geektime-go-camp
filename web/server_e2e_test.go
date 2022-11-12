//go:build e2e

package web

import "testing"

func TestHttpServer(t *testing.T) {
	s := NewHttpServer()

	s.Get("/", func(ctx *Context) {
		ctx.Res.Write([]byte("hello, world"))
	})

	err := s.Start(":8081")
	if err != nil {
		return
	}
}
