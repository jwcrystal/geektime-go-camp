package sso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"geektime-go/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"io"
	"net/http"
	"testing"
	"time"
)

// 使用 Redis 作為 session, 瓶頸優化
var bSessions = cache.New(15*time.Minute, time.Second)

// 在業務服務器上，模擬單機登入過程
func testBizServer_B(t *testing.T) {
	server := web.NewHttpServer(web.ServerWithMiddlewares(LoginMiddlewareServerB))
	// 需要登入才看得到，如何處理？
	// 要判斷是否有登入， 這邊透過 middleware 進行登入檢驗

	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespJSONOK(&User{
			Name: "Tom B",
			Age:  18,
		})
	})

	server.Get("/token", func(ctx *web.Context) {
		token, err := ctx.QueryValue("token")
		if err != nil {
			_ = ctx.RespServerError("Biz server - B: token 不對")
			return
		}
		signature := Encrypt("server_b")
		req, err := http.NewRequest(http.MethodPost,
			"http://sso.com:8080/token/validate?token="+token,
			bytes.NewBuffer([]byte(signature)))
		if err != nil {
			_ = ctx.RespServerError("Biz server - B: 解析 token 失敗")
			return
		}
		t.Log(req)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			_ = ctx.RespServerError("Biz server - B: 解析 token 失敗")
			return
		}
		tokenBs, _ := io.ReadAll(resp.Body)
		var tokens Tokens
		// 會拿到兩個 tokens
		// 1. access token
		// 2. refresh token
		_ = json.Unmarshal(tokenBs, &tokens)

		sessionId := uuid.New().String()
		bSessions.Set(sessionId, tokens, 15*time.Minute)
		ctx.SetCookie(&http.Cookie{
			Name:  "b_sessid",
			Value: sessionId,
		})

		// 登入成功，跳回最一開始的 /profile 頁面
		http.Redirect(ctx.Resp, ctx.Req, "http://bbb.com:8082/profile", http.StatusFound)
	})
	err := server.Start(":8082")
	t.Log(err)
}

// 登入驗證的 Middleware
func LoginMiddlewareServerB(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/token" {
			next(ctx)
			return
		}
		// ssid，即 session id
		redirect := fmt.Sprintf("http://sso.com:8080/login?client_id=server_b")
		cookie, err := ctx.Req.Cookie("b_sessid")
		if err != nil {
			// 登入失敗
			//ctx.RespServerError("Biz server - B: 沒有登入 token")
			http.Redirect(ctx.Resp, ctx.Req, redirect, http.StatusFound)
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		_, ok := bSessions.Get(ssid)
		if !ok {
			// 登入失敗
			ctx.RespServerError("Biz server - B: 沒有登入 session id")
			return
		}

		next(ctx)
	}
}
