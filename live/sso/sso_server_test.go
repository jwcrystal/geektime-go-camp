package sso

import (
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
var ssoSession = cache.New(time.Minute*15, time.Second)

// 模擬單機登入，把 SSO 驗證服務器 與 業務服務器 放一起
func TestSSOServer(t *testing.T) {
	whiteList := map[string]string{
		"server_a": "http://aaa.com:8081/token",
		"server_b": "http://bbb.com:8082/token",
	}
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{T: tpl}

	server := web.NewHttpServer(web.ServerWithTemplateEngine(engine))
	server.Get("/login", func(ctx *web.Context) {
		// 要判斷是否有登入， 這邊透過 middleware 進行登入檢驗

		ck, err := ctx.Req.Cookie("token")
		clientId, _ := ctx.QueryValue("client_id")
		if err != nil {
			_ = ctx.Render("login.gohtml",
				map[string]string{"ClientId": clientId})
			//_ = ctx.Render("login.gohtml", map[string]string{"ClientId": clientId})
			return
		}

		// 如果 client_id 和已有 session 歸屬不同主題，還是需要重新登入
		// 建立一個 client_id 到 session 的映射
		_, ok := ssoSession.Get(ck.Value)
		if !ok {
			_ = ctx.Render("login.gohtml",
				map[string]string{"ClientId": clientId})
			return
		}

		// token
		token := uuid.New().String()
		ssoSession.Set(clientId, token, time.Minute)
		http.Redirect(ctx.Resp, ctx.Req, whiteList[clientId]+"?token="+token, http.StatusFound)
	})

	// 模擬登入, login.gohtml
	server.Post("/login", func(ctx *web.Context) {
		if err != nil {
			ctx.RespServerError("SSO server: 系統錯誤")
			return
		}
		// Verify data
		email, _ := ctx.FormValue("email")
		password, _ := ctx.FormValue("password")
		clientId, _ := ctx.FormValue("client_id")

		// hardcode, 只檢測這個例子
		if email == "abc@biz.com" && password == "123" {
			// login successfully
			// 如果要防止 token 被盜走，不能使用 uuid
			id := uuid.New().String()
			http.SetCookie(ctx.Resp, &http.Cookie{
				Name:    "token",
				Value:   id,
				Expires: time.Now().Add(15 * time.Minute),
			})
			ssoSession.Set(id, &User{Name: "Tom"}, 15*time.Minute)
			token := uuid.New().String()
			ssoSession.Set(clientId, token, 15*time.Minute)
			http.Redirect(ctx.Resp, ctx.Req, whiteList[clientId]+"?token="+token, http.StatusFound)
			return
		}
		ctx.RespServerError("SSO server: 用戶帳號密碼不對")
	})

	// 驗證 token， 如何提供？
	// 1. 頻率限制
	// 2. 來源檢驗
	server.Post("/token/validate", func(ctx *web.Context) {
		token, err := ctx.QueryValue("token")
		if err != nil {
			_ = ctx.RespServerError("SSO server: 拿不到 token")
			return
		}
		signature, err := io.ReadAll(ctx.Req.Body)
		if err != nil {
			_ = ctx.RespServerError("SSO server: 拿不到 簽名驗證")
			return
		}
		clientId, _ := Decrypt(signature)

		val, ok := ssoSession.Get(clientId)
		if !ok {
			// client_id 沒登入或是 token 過期了
			_ = ctx.RespServerError("SSO server: 沒登入")
			return
		}
		if token != val {
			_ = ctx.RespServerError("SSO server: token 不對")
			return
		}
		// 認證碼 code or token 只能用一次
		ssoSession.Delete(clientId)

		_ = ctx.RespJSONOK(Tokens{
			AccessToken:  uuid.New().String(),
			RefreshToken: uuid.New().String(),
		})
	})

	// server a
	go func() {
		testBizServer_A(t)
	}()

	// server b
	go func() {
		testBizServer_B(t)
	}()

	err = server.Start(":8080")
	t.Log(err)
}

type User struct {
	Name     string
	Password string
	Age      int
}

type Tokens struct {
	AccessToken           string  `json:"access_token"`
	AccessTokenExpiration float64 `json:"access_token_expiration"`
	RefreshToken          string  `json:"refresh_token"`
}
