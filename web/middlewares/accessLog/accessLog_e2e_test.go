package accessLog

import (
	"fmt"
	"geektime-go/web"
	"testing"
)

func TestMiddlewareBuilderE2E(t *testing.T) {
	builder := MiddlewareBuilder{}
	builder.LogFunc(func(log string) {
		fmt.Println(log)
	})
	h := web.NewHttpServerV2(web.ServerWithMiddlewares(builder.Build()))
	// another way
	//mdls := builder.LogFunc(func(log string) {
	//	fmt.Println(log)
	//}).Build()
	//h := web.NewHttpServerV2(web.ServerWithMiddlewares(mdls))

	h.Get("/", func(ctx *web.Context) {
		ctx.Res.Write([]byte("Hello accessLog"))
	})
	h.Get("/user", func(ctx *web.Context) {
		ctx.Res.Write([]byte("Hello, user"))
	})

	h.Start(":8081")
}
