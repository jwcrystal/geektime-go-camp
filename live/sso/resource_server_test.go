package sso

import (
	"geektime-go/web"
	"io"
	"net/http"
	"testing"
)

func testBizServer_Resource(t *testing.T) {
	server := web.NewHttpServer()
	server.Get("/profile", func(ctx *web.Context) {
		token, _ := ctx.QueryValue("token")
		// 可能是 RPC 調用，因為 授權服務 和 資源服務，都是同一個公司的
		req, _ := http.NewRequest("POST", "http://auth.com:8080/token/validate?token="+token, nil)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			_ = ctx.RespServerError("Resource server - GeekTime: token 不對")
			return
		}
		data, _ := io.ReadAll(resp.Body)
		// 檢驗 scope
		if string(data) != "basic" {
			_ = ctx.RespServerError("Resource server - GeekTime: 沒有權限 ( " + string(data) + " )")
			return
		}
		_ = ctx.RespJSONOK(User{
			Name: "Jared Wang",
		})
	})
	err := server.Start(":8082")
	t.Log(err)
}
