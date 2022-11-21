package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_AddRoute(t *testing.T) {
	// first, build the route tree
	// second, verify the tree
	testRoutes := []struct {
		method string
		path   string
	}{
		// static route
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		// 通配符測試用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/aaa",
		},
		{
			method: http.MethodGet,
			path:   "/*/aaa/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 參數路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// Regexp 路由 （正則路由）
		{
			method: http.MethodPatch,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodPatch,
			path:   "/:user(^.+$)/hello",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		// 測試route，不需要理會handler處理
		r.addRoute(route.method, route.path, mockHandler)
	}

	// 斷言 route tree 跟預期的一樣
	wantRouter := &Router{
		trees: map[string]*node{
			http.MethodGet: {
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": {
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": {
								path:     "home",
								handler:  mockHandler,
								nodeType: nodeTypeStatic,
							},
						},
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {
								path:     "detail",
								handler:  mockHandler,
								nodeType: nodeTypeStatic,
							},
						},
						starChild: &node{
							path:     "*",
							handler:  mockHandler,
							nodeType: nodeTypeAny,
						},
					},
					"param": {
						path: "param",
						paramChild: &node{
							path:        ":id",
							paramString: "id",
							handler:     mockHandler,
							nodeType:    nodeTypeParam,
							children: map[string]*node{
								"detail": {
									path:     "detail",
									handler:  mockHandler,
									nodeType: nodeTypeStatic,
								},
							},
							starChild: &node{path: "*", handler: mockHandler, nodeType: nodeTypeAny},
						},
					},
				},
				starChild: &node{
					path:     "*",
					handler:  mockHandler,
					nodeType: nodeTypeAny,
					children: map[string]*node{
						"aaa": {
							path: "aaa",
							starChild: &node{
								path:     "*",
								handler:  mockHandler,
								nodeType: nodeTypeAny,
							},
							handler:  mockHandler,
							nodeType: nodeTypeStatic,
						},
					},
					starChild: &node{
						path:     "*",
						handler:  mockHandler,
						nodeType: nodeTypeAny,
					},
				},
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {
						path: "order",
						children: map[string]*node{
							"create": {
								path:     "create",
								handler:  mockHandler,
								nodeType: nodeTypeStatic,
							},
						},
					},
					"login": {
						path:     "login",
						handler:  mockHandler,
						nodeType: nodeTypeStatic,
					},
				},
			},
			http.MethodPatch: {
				path: "/",
				children: map[string]*node{
					"reg": {
						path: "reg",
						regChild: &node{
							path:        ":id(.*)",
							paramString: "id",
							handler:     mockHandler,
							//children: map[string]*node{},
							nodeType: nodeTypeRegexp,
						},
					},
				},
				regChild: &node{
					path:        ":user(^.+$)",
					paramString: ":user",
					nodeType:    nodeTypeRegexp,
					children: map[string]*node{
						"hello": {
							path:    "hello",
							handler: mockHandler,
						},
					},
				},
			},
		},
	}

	// 因為 HandleFunc 不可比較
	//assert.Equal(t, wantRouter, r)
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	// 處理abnormal route （非法路由）
	r = NewRouter()
	// empty path
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	}, "web: not start with '/'")
	// not start with "/"
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "login", mockHandler)
	}, "web: not start with '/'")
	// end with "/"
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/login/", mockHandler)
	}, "web: end with '/'")
	// Continuous "//"
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "//login", mockHandler)
	}, "web: no continuous '//' ")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	}, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	}, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]")

	r = NewRouter()
	r.addRoute(http.MethodGet, "/", mockHandler)
	// root node register twice
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	}, "web: the route conflicts, '/' register twice")
	// node register twice
	r = NewRouter()
	assert.PanicsWithValue(t, fmt.Sprintf("web: the route conflicts, /a/b/c register twice"), func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})
	// parameter and star path verification
	r = NewRouter()
	assert.PanicsWithValue(t, fmt.Sprintf("web: invalid route, a parameter path existed"), func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
	})
	r = NewRouter()
	r.addRoute(http.MethodGet, "/a/*", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	}, "web: invalid route, a star path existed")
	r = NewRouter()
	// 参数冲突
	r.addRoute(http.MethodGet, "/a/:user", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	}, "web: invalid route, parameter route conflict")
	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	}, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]")
	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	}, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]")
	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	}, "web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [:id(.*)]")
	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
	}, "web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [:id]")
	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	}, "web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [*]")
	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	}, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [:id(.*)]")
}

func (r *Router) equal(y *Router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("invalid http method"), false
		}
		msg, ok := v.equal(dst)
		if !ok {
			return msg, false
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return fmt.Sprintf("Node path %s not found", y.path), false
	}
	if n.path != y.path {
		return fmt.Sprintf("%s: the node path does not matched x: %s, y :%s", n.path, n.path, y.path), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s: the number of children was not matched", n.path), false
	}

	if n.nodeType != y.nodeType {
		return fmt.Sprintf("%s: not same node type x: %d, y: %d", n.path, n.nodeType, y.nodeType), false
	}

	if len(n.children) == 0 {
		return "", true
	}

	if n.starChild != nil {
		msg, ok := n.starChild.equal(y.starChild)
		if !ok {
			return msg, ok
		}
	}

	if n.paramChild != nil {
		msg, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return msg, ok
		}
	}
	if n.regChild != nil {
		msg, ok := n.regChild.equal(y.regChild)
		if !ok {
			return msg, ok
		}
	}
	// 比較 handler
	nhandler := reflect.ValueOf(n.handler)
	yhandler := reflect.ValueOf(y.handler)

	if nhandler != yhandler {
		return fmt.Sprintf("%s: handler was not equal x: %s, y: %s", n.path, nhandler.Type().String(), yhandler.Type().String()), false
	}

	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("%s: sub-node %s not existed", n.path, path), false
		}
		msg, ok := c.equal(dst)
		if !ok {
			return msg, ok
		}
	}
	return "", true
}

func TestRouter_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		// static route
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 參數路由
		{
			method: http.MethodPost,
			path:   "/param/:id",
		},
		{
			method: http.MethodPost,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodPost,
			path:   "/param/:id/*",
		},
		//// 通配符測試用例
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		// 正則路由測試用例
		{
			method: http.MethodPatch,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodPatch,
			path:   "/:id(^[0-9]+$)/hello",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		matchInfo *matchInfo
	}{
		{
			// method does not exist
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:      "root",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			// 完全命中
			name:      "user home",
			method:    http.MethodGet,
			path:      "/user/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 完全命中
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "detail",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "no handler",
			method:    http.MethodGet,
			path:      "/order",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path: "order",
					children: map[string]*node{
						"detail": {
							path:    "detail",
							handler: mockHandler,
						},
					},
				},
			},
		},
		{
			name:      "two layer",
			method:    http.MethodPost,
			path:      "/order/create",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中 /order/*, * matched
			name:      "star matched",
			method:    http.MethodPost,
			path:      "/order/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中 /order/*/home, * matched
			name:      "star in the middle",
			method:    http.MethodGet,
			path:      "/user/user/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段，支持末尾通配一段
			name:      "overflow",
			method:    http.MethodPost,
			path:      "/order/delete/123",
			wantFound: true,
			matchInfo: &matchInfo{ //支持通配末尾多段
				node: &node{
					path:     "*",
					handler:  mockHandler,
					nodeType: nodeTypeAny,
				},
			},
		},
		{
			// 參數路由
			name:      ":id",
			method:    http.MethodPost,
			path:      "/param/123",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"id": "123",
				},
			},
		},
		{
			// 命中 /param/:id/*
			name:      ":id*",
			method:    http.MethodPost,
			path:      "/param/123/abc",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/detail
			name:      ":id*",
			method:    http.MethodPost,
			path:      "/param/123/detail",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 正則路由
			name:      "RegExpr id(.*)",
			method:    http.MethodPatch,
			path:      "/reg/abc",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"id": "abc",
				},
			},
		},
		{
			name:      "RegExpr id(^[0-9]+$)",
			method:    http.MethodPatch,
			path:      "/123",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path: ":id(^[0-9]+$)",
					children: map[string]*node{
						"hello": {
							path:    "hello",
							handler: mockHandler,
						},
					},
					nodeType: nodeTypeRegexp,
				},
				pathParams: map[string]string{
					"id": "123",
				},
			},
		},
		{
			// 未命中 /id(^[0-9]+$)/home
			name:   "RegExpr not :id(^[0-9]+$)",
			method: http.MethodDelete,
			path:   "/abc/home",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			route, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			// handler 無法比較
			//assert.Equal(t, tc.matchInfo.pathParams, route.pathParams)
			//msg, ok := tc.matchInfo.node.equal(route.node)
			//assert.True(t, ok, msg)
			assert.Equal(t, tc.matchInfo.pathParams, route.pathParams)
			n := tc.matchInfo.node
			wantVal := reflect.ValueOf(tc.matchInfo.node.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}

func Benchmark_findRoute_Static(t *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		// static route
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
	}
	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}
	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		matchInfo *matchInfo
	}{
		{
			name:      "root",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			// 完全命中
			name:      "user home",
			method:    http.MethodGet,
			path:      "/user/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 完全命中
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "detail",
					handler: mockHandler,
				},
			},
		},
	}
	t.ResetTimer()
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func Benchmark_findRoute_Any(t *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
	}
	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}
	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		matchInfo *matchInfo
	}{
		// 通配符匹配
		{
			// 命中/order/*
			name:      "star match",
			method:    http.MethodPost,
			path:      "/order/delete",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中通配符在中间的
			// /user/*/home
			name:      "star in middle",
			method:    http.MethodGet,
			path:      "/user/Tom/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
	}
	t.ResetTimer()
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func Benchmark_findRoute_Param(t *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
	}
	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}
	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		matchInfo *matchInfo
	}{
		// 参数匹配
		{
			// 命中 /param/:id
			name:      ":id",
			method:    http.MethodGet,
			path:      "/param/123",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/*
			name:      ":id*",
			method:    http.MethodGet,
			path:      "/param/123/abc",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/detail
			name:      ":id*",
			method:    http.MethodGet,
			path:      "/param/123/detail",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
	}
	t.ResetTimer()
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func Benchmark_findRoute_RegExpr(t *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		// 正则
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:id([0-9]+)/home",
		},
	}
	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}
	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		matchInfo *matchInfo
	}{
		{
			// 命中 /reg/:id(.*)
			name:      ":id(.*)",
			method:    http.MethodDelete,
			path:      "/reg/123",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /:id([0-9]+)/home
			name:      ":id([0-9]+)",
			method:    http.MethodDelete,
			path:      "/123/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
	}
	t.ResetTimer()
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func TestRouter_findRoute_Middleware(t *testing.T) {
	var mdlBuilder = func(i byte) Middleware {
		return func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				ctx.ResData = append(ctx.ResData, i)
				next(ctx)
			}
		}
	}
	mdlsRoute := []struct {
		method string
		path   string
		mdls   []Middleware
	}{
		{
			method: http.MethodGet,
			path:   "/a/b",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b')},
		},
		{
			method: http.MethodGet,
			path:   "/a/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*')},
		},
		{
			method: http.MethodGet,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/*/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*'), mdlBuilder('c')},
		},
		{
			method: http.MethodPost,
			path:   "/a/b/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('c')},
		},
		{
			method: http.MethodDelete,
			path:   "/*",
			mdls:   []Middleware{mdlBuilder('*')},
		},
		{
			method: http.MethodDelete,
			path:   "/",
			mdls:   []Middleware{mdlBuilder('/')},
		},
	}
	r := NewRouter()
	for _, mdlRoute := range mdlsRoute {
		r.addRoute(mdlRoute.method, mdlRoute.path, nil, mdlRoute.mdls...)
	}
	testCases := []struct {
		name   string
		method string
		path   string
		// 我们借助 ctx 里面的 RespData 字段来判断 middleware 有没有按照预期执行
		wantResp string
	}{
		{
			name:   "static, not match",
			method: http.MethodGet,
			path:   "/a",
		},
		{
			name:     "static, match",
			method:   http.MethodGet,
			path:     "/a/c",
			wantResp: "a*",
		},
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b",
			wantResp: "a*ab",
		},
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b/c",
			wantResp: "a*abab*",
		},
		{
			name:     "abc",
			method:   http.MethodPost,
			path:     "/a/b/c",
			wantResp: "a*cab*abc",
		},
		{
			name:     "root",
			method:   http.MethodDelete,
			path:     "/",
			wantResp: "/",
		},
		{
			name:     "root star",
			method:   http.MethodDelete,
			path:     "/a",
			wantResp: "/*",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, _ := r.findRouteWithMiddleware(tc.method, tc.path)
			mdls := mi.middlewares
			var root HandleFunc = func(ctx *Context) {
				// 使用 string 可读性比较高
				assert.Equal(t, tc.wantResp, string(ctx.ResData))
			}
			for i := len(mdls) - 1; i >= 0; i-- {
				root = mdls[i](root)
			}
			// 开始调度
			root(&Context{
				ResData: make([]byte, 0, len(tc.wantResp)),
			})
		})
	}
}

func Benchmark_findRoute1(b *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		// static route
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		//// 通配符測試用例
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		matchInfo *matchInfo
	}{
		{
			name:      "root",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			// 完全命中
			name:      "user home",
			method:    http.MethodGet,
			path:      "/user/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 完全命中
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "detail",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "no handler",
			method:    http.MethodGet,
			path:      "/order",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path: "order",
					children: map[string]*node{
						"detail": {
							path:    "detail",
							handler: mockHandler,
						},
					},
				},
			},
		},
		{
			name:      "two layer",
			method:    http.MethodPost,
			path:      "/order/create",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中 /order/*, * matched
			name:      "star matched",
			method:    http.MethodPost,
			path:      "/order/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中 /order/*/home, * matched
			name:      "star in the middle",
			method:    http.MethodGet,
			path:      "/user/user/home",
			wantFound: true,
			matchInfo: &matchInfo{
				node: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段，支持末尾通配一段
			name:      "overflow",
			method:    http.MethodPost,
			path:      "/order/delete/123",
			wantFound: true,
			matchInfo: &matchInfo{ //支持通配末尾多段
				node: &node{
					path:     "*",
					handler:  mockHandler,
					nodeType: nodeTypeAny,
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func Benchmark_findRoute1_Middleware(b *testing.B) {
	var mdlBuilder = func(i byte) Middleware {
		return func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				ctx.ResData = append(ctx.ResData, i)
				next(ctx)
			}
		}
	}
	mdlsRoute := []struct {
		method string
		path   string
		mdls   []Middleware
	}{
		{
			method: http.MethodGet,
			path:   "/a/b",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b')},
		},
		{
			method: http.MethodGet,
			path:   "/a/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*')},
		},
		{
			method: http.MethodGet,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/*/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*'), mdlBuilder('c')},
		},
		{
			method: http.MethodPost,
			path:   "/a/b/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('c')},
		},
		{
			method: http.MethodDelete,
			path:   "/*",
			mdls:   []Middleware{mdlBuilder('*')},
		},
		{
			method: http.MethodDelete,
			path:   "/",
			mdls:   []Middleware{mdlBuilder('/')},
		},
	}
	r := NewRouter()
	for _, mdlRoute := range mdlsRoute {
		r.addRoute(mdlRoute.method, mdlRoute.path, nil, mdlRoute.mdls...)
	}
	testCases := []struct {
		name   string
		method string
		path   string
		// 我们借助 ctx 里面的 RespData 字段来判断 middleware 有没有按照预期执行
		wantResp string
	}{
		{
			name:   "static, not match",
			method: http.MethodGet,
			path:   "/a",
		},
		{
			name:     "static, match",
			method:   http.MethodGet,
			path:     "/a/c",
			wantResp: "a*",
		},
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b",
			wantResp: "a*ab",
		},
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b/c",
			wantResp: "a*abab*",
		},
		{
			name:     "abc",
			method:   http.MethodPost,
			path:     "/a/b/c",
			wantResp: "a*cab*abc",
		},
		{
			name:     "root",
			method:   http.MethodDelete,
			path:     "/",
			wantResp: "/",
		},
		{
			name:     "root star",
			method:   http.MethodDelete,
			path:     "/a",
			wantResp: "/*",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			r.findRouteWithMiddleware(tc.method, tc.path)
		}
	}
}
