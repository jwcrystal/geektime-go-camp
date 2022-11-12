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
			path:   "/user/:id",
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
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {
								path:    "detail",
								handler: mockHandler,
							},
						},
						starChild: &node{
							path:    "*",
							handler: mockHandler,
						},
					},
				},
				starChild: &node{
					path:    "*",
					handler: mockHandler,
					children: map[string]*node{
						"aaa": {
							path: "aaa",
							starChild: &node{
								path:    "*",
								handler: mockHandler,
							},
							handler: mockHandler,
						},
					},
					starChild: &node{
						path:    "*",
						handler: mockHandler,
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
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": {
						path:    "login",
						handler: mockHandler,
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

	r = NewRouter()
	r.addRoute(http.MethodGet, "/", mockHandler)
	// root node register twice
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	}, "web: the route conflicts, '/' register twice")
	// node register twice
	r = NewRouter()
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	// root node register twice
	assert.PanicsWithValue(t, fmt.Sprintf("web: the route conflicts, /a/b/c register twice"), func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)

	})
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
		return fmt.Sprintf("the node path does not matched"), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("the number of children was not matched"), false
	}

	if n.starChild != nil {
		msg, ok := n.starChild.equal(y.starChild)
		if !ok {
			return msg, ok
		}
	}

	// 比較 handler
	nhandler := reflect.ValueOf(n.handler)
	yhandler := reflect.ValueOf(y.handler)

	if nhandler != yhandler {
		return fmt.Sprintf("handler was not equal"), false
	}

	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("sub-node %s not existed", path), false
		}
		msg, ok := c.equal(dst)
		if !ok {
			return msg, false
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
		//{
		//	method: http.MethodGet,
		//	path:   "/*",
		//},
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
		wantNode  *node
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
			method:    http.MethodGet,
			path:      "/",
			wantFound: true,
			wantNode: &node{
				path:    "/",
				handler: mockHandler,
			},
		},
		{
			// 完全命中
			name:      "user home",
			method:    http.MethodGet,
			path:      "/user/home",
			wantFound: true,
			wantNode: &node{
				path:    "home",
				handler: mockHandler,
			},
		},
		{
			// 完全命中
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			wantNode: &node{
				path:    "detail",
				handler: mockHandler,
			},
		},
		{
			name:      "no handler",
			method:    http.MethodPost,
			path:      "/order",
			wantFound: true,
			wantNode: &node{
				path: "order",
			},
		},
		{
			name:      "two layer",
			method:    http.MethodPost,
			path:      "/order/create",
			wantFound: true,
			wantNode: &node{
				path:    "create",
				handler: mockHandler,
			},
		},
		{
			// 命中 /order/*, * matched
			name:      "star matched",
			method:    http.MethodPost,
			path:      "/order/home",
			wantFound: true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			// 命中 /order/*/home, * matched
			name:      "star in the middle",
			method:    http.MethodGet,
			path:      "/user/user/home",
			wantFound: true,
			wantNode: &node{
				path:    "home",
				handler: mockHandler,
			},
		},
		{
			// 比 /order/* 多了一段
			name:   "overflow",
			method: http.MethodPost,
			path:   "/order/delete/123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			// handler 無法比較
			//assert.Equal(t, tc.wantNode, n)
			//msg, ok := tc.wantNode.equal(n)
			//assert.True(t, ok, msg)
			assert.Equal(t, tc.wantNode.path, n.path)
			wantVal := reflect.ValueOf(tc.wantNode.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}
