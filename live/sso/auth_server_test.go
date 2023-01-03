package sso

import (
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
// var authSessions = map[string]any{}
var authSessions = cache.New(time.Minute*15, time.Second)

// 模擬單機登入，把 SSO 驗證服務器 與 業務服務器 放一起
func TestAuthServer(t *testing.T) {
	whiteList := map[string]string{
		"server_geek": "http://geek.com:8081/callback?code=",
		"server_a":    "http://aaa.com:8081/token",
		"server_b":    "http://bbb.com:8082/token",
	}
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{T: tpl}

	server := web.NewHttpServer(web.ServerWithTemplateEngine(engine))
	// confirm.gohtml
	server.Get("/auth", func(ctx *web.Context) {
		clientId, _ := ctx.QueryValue("client_id")
		scope, _ := ctx.QueryValue("scope")
		_ = ctx.Render("confirm.gohtml",
			map[string]string{"ClientId": clientId, "Scope": scope})
	})

	// 模擬登入, login.gohtml
	server.Post("/auth", func(ctx *web.Context) {
		if err != nil {
			ctx.RespServerError("Auth server: 系統錯誤")
			return
		}
		// 校驗資料
		clientId, _ := ctx.QueryValue("client_id")
		scope, _ := ctx.QueryValue("scope")
		fmt.Println(clientId, scope)
		// 授權碼 code
		code := uuid.New().String()
		authSessions.Set(code, map[string]string{
			"client_id": clientId,
			"scope":     scope,
		}, time.Minute*15)
		http.Redirect(ctx.Resp, ctx.Req, whiteList[clientId]+code, http.StatusFound)
	})

	// 驗證 token， 如何提供？
	// 1. 頻率限制
	// 2. 來源檢驗
	server.Post("/token", func(ctx *web.Context) {
		code, err := ctx.QueryValue("code")
		if err != nil {
			_ = ctx.RespServerError("Auth server: 拿不到 code")
			return
		}
		signature, err := io.ReadAll(ctx.Req.Body)
		if err != nil {
			_ = ctx.RespServerError("Auth server: 拿不到 簽名驗證")
			return
		}
		clientId, _ := Decrypt(signature)

		val, ok := authSessions.Get(code)
		if !ok {
			// 沒有 code 或是 token 過期了
			_ = ctx.RespServerError("Auth server: 非法 code")
			return
		}
		// 不是單一值，無法直接比較
		//if code != val {
		//	_ = ctx.RespServerError("Auth server: code 不對")
		//	return
		//}
		data := val.(map[string]string)
		codeClientId := data["client_id"]
		if clientId != codeClientId {
			_ = ctx.RespServerError("Auth server: code 不對，有人劫持了 code")
			return
		}
		// 認證碼 code or token 只能用一次
		authSessions.Delete(code)

		// 產生訪問資源的 token
		accessToken := uuid.New().String()
		// 建立一個 access token 到 scope（訪問權限） 的映射
		authSessions.Set(accessToken, data["scope"], time.Minute*15)
		// 返回 雙 token：
		// access token + refresh token
		_ = ctx.RespJSONOK(Tokens{
			AccessToken:           accessToken,
			AccessTokenExpiration: (15 * time.Minute).Seconds(),
			RefreshToken:          uuid.New().String(),
		})
	})

	server.Post("/token/validate", func(ctx *web.Context) {
		token, _ := ctx.QueryValue("token")
		scope, ok := authSessions.Get(token)
		if !ok {
			_ = ctx.RespServerError("Auth server: 非法 token")
			return
		}
		_ = ctx.RespString(http.StatusOK, scope.(string))
	})

	// server a
	go func() {
		testBizServer_GeekTime(t)
	}()

	// server b
	go func() {
		testBizServer_Resource(t)
	}()

	err = server.Start(":8080")
	t.Log(err)
}
