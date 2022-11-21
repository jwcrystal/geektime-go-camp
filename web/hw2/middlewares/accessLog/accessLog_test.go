package accessLog

import (
	"fmt"
	"geektime-go/web"
	"net/http"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{}
	builder.LogFunc(func(log string) {
		fmt.Println(log)
	})
	h := web.NewHttpServer()
	//h := web.NewHttpServerV2(web.ServerWithMiddlewares(builder.Build()))
	// another way
	//mdls := builder.LogFunc(func(log string) {
	//	fmt.Println(log)
	//}).Build()
	//h := web.NewHttpServerV2(web.ServerWithMiddlewares(mdls))

	h.Get("/user", func(ctx *web.Context) {
		fmt.Println("hello" + ctx.Req.URL.Path)
	})
	h.Post("/a/b/*", func(ctx *web.Context) {
		fmt.Println("hello post")
	})

	h.UseV1(http.MethodGet, "/test", func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			fmt.Println("hello, test")
			next(ctx)
		}
	})
	//log := func(next web.HandleFunc) web.HandleFunc {
	//	return func(ctx *web.Context) {
	//		fmt.Println(&ctx.MatchedRoute)
	//		next(ctx)
	//	}
	//}
	//h.UseV1("Get", "/", log)
	//h.UseV1("Post", "/user", log)
	//req, err := http.NewRequest(http.MethodPost, "/a/b/c", nil)
	//if err != nil {
	//	t.Fatal(err)
	//}
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(nil, req)
}
