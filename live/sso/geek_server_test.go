package sso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"geektime-go/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	"html/template"
	"io"
	"net/http"
	"testing"
	"time"
)

// 使用 Redis 作為 session, 瓶頸優化
var geekSessions = cache.New(15*time.Minute, time.Second)

// 在業務服務器上，模擬單機登入過程
func testBizServer_GeekTime(t *testing.T) {
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{T: tpl}
	server := web.NewHttpServer(
		web.ServerWithTemplateEngine(engine),
		web.ServerWithMiddlewares(LoginMiddlewareServerGeekTime))
	// 需要登入才看得到，如何處理？
	// 要判斷是否有登入， 這邊透過 middleware 進行登入檢驗

	server.Get("/home", func(ctx *web.Context) {
		cookie, err := ctx.Req.Cookie("geek_sessid")
		if err != nil {
			_ = ctx.RespServerError("Biz server - GeekTime: 服務器錯誤")
			return
		}

		val, _ := geekSessions.Get(cookie.Value)
		ctx.RespString(http.StatusOK, "Hello, "+val.(User).Name)
	})

	server.Get("/callback", func(ctx *web.Context) {
		code, err := ctx.QueryValue("code")
		if err != nil {
			_ = ctx.RespServerError("Biz server - GeekTime: code 不對")
			return
		}
		signature := Encrypt("server_geek")
		req, err := http.NewRequest(http.MethodPost,
			"http://auth.com:8080/token?code="+code,
			bytes.NewBuffer([]byte(signature)))
		if err != nil {
			_ = ctx.RespServerError("Biz server - GeekTime: 解析 code 失敗")
			return
		}
		t.Log(req)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			_ = ctx.RespServerError("Biz server - GeekTime: 解析 code 失敗")
			return
		}
		tokenBs, _ := io.ReadAll(resp.Body)
		var tokens Tokens
		// 會拿到兩個 tokens
		// 1. access token
		// 2. refresh token
		_ = json.Unmarshal(tokenBs, &tokens)

		// 正常 access token 要小心被竊取，所以請求還會帶上 client id （對應 app id） 或是 state 參數
		// 用 token 去訪問 資源服務器
		req, err = http.NewRequest(http.MethodGet,
			"http://resource.com:8082/profile?token="+tokens.AccessToken,
			nil)
		if err != nil {
			_ = ctx.RespServerError("Biz server - GeekTime: 解析 code 失敗")
			return
		}
		t.Log(req)

		resp, err = (&http.Client{}).Do(req)
		if err != nil {
			_ = ctx.RespServerError("Biz server - GeekTime: 資源服務器 異常")
			return
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			_ = ctx.RespServerError("Biz server - GeekTime: 資源服務器 異常")
			return
		}
		var u User
		_ = json.Unmarshal(respBody, &u)
		// 登入成功
		sessionId := uuid.New().String()
		geekSessions.Set(sessionId, u, 15*time.Minute)
		ctx.SetCookie(&http.Cookie{
			Name:  "geek_sessid",
			Value: sessionId,
		})

		// 登入成功，跳回最一開始的 /home 頁面
		http.Redirect(ctx.Resp, ctx.Req, "http://geek.com:8081/home", http.StatusFound)
	})
	err = server.Start(":8081")
	t.Log(err)
}

// 登入驗證的 Middleware
func LoginMiddlewareServerGeekTime(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/callback" {
			next(ctx)
			return
		}
		redirect := fmt.Sprintf("http://auth.com:8080/auth?client_id=server_geek&scope=basic")
		// ssid，即 session id
		cookie, err := ctx.Req.Cookie("geek_sessid")
		if err != nil {
			// 登入失敗
			//ctx.RespServerError("Biz server - GeekTime: 沒有登入 token")
			//http.Redirect(ctx.Resp, ctx.Req, redirect, http.StatusFound)
			_ = ctx.Render("login_geek.gohtml", map[string]string{"RedirectURL": redirect})
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		_, ok := geekSessions.Get(ssid)
		if !ok {
			// 登入失敗
			//ctx.RespServerError("Biz server - GeekTime: 沒有登入 session id")
			_ = ctx.Render("login_geek.gohtml", map[string]string{"RedirectURL": redirect})
			return
		}
		// 驗證成功，登入
		next(ctx)
	}
}
