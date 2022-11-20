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
	h := web.NewHttpServerV2(web.ServerWithMiddlewares(builder.Build()))
	// another way
	//mdls := builder.LogFunc(func(log string) {
	//	fmt.Println(log)
	//}).Build()
	//h := web.NewHttpServerV2(web.ServerWithMiddlewares(mdls))

	h.Post("/a/b/*", func(ctx *web.Context) {
		fmt.Println("hello post")
	})
	//h.Use("Post", "/")
	req, err := http.NewRequest(http.MethodGet, "/a/b/c", nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(nil, req)
}
