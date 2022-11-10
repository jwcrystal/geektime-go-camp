package web

import "testing"

func TestHttpServer(t *testing.T) {
	s := &HttpServer{}

	s.Get("/", func(ctx *Context) {
		ctx.res.Write([]byte("hello, world"))
	})

	err := s.Start(":8080")
	if err != nil {
		return
	}
}
