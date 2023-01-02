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
	//h := web.NewHttpServer()
	// another way
	//mdls := builder.LogFunc(func(log string) {
	//	fmt.Println(log)
	//}).Build()
	//h := web.NewHttpServerV2(web.ServerWithMiddlewares(mdls))
	log := func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			fmt.Println(&ctx.MatchedRoute)
			ctx.Resp.Write([]byte(ctx.MatchedRoute))
			next(ctx)
		}
	}
	h.UseV1("Get", "/", log)
	h.UseV1("Get", "/user", func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			fmt.Println(&ctx.MatchedRoute)
			ctx.Resp.Write([]byte(ctx.MatchedRoute + " 1"))
			next(ctx)
		}
	})

	//h.Get("/", func(ctx *web.Context) {
	//	ctx.Resp.Write([]byte("Hello accessLog"))
	//})
	//h.Get("/user", func(ctx *web.Context) {
	//	ctx.Resp.Write([]byte("Hello, user"))
	//})

	h.Start(":8081")
}
