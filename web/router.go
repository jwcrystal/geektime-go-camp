package web

import (
	"fmt"
	"strings"
)

type node struct {
	path string
	// sub-node
	children map[string]*node

	// 通配符 "*"
	starChild *node

	// 參數路由
	paramChild *node

	// non-core operation
	handler HandleFunc
}

// Router tree
type Router struct {
	trees map[string]*node
}

type matchInfo struct {
	node       *node
	pathParams map[string]string
}

func NewRouter() *Router {
	return &Router{trees: map[string]*node{}}
}

// addRoute path must start with "/", not end with "/", not continues with "//"
func (r *Router) addRoute(method string, path string, handlerFunc HandleFunc) {
	if path == "" {
		panic("web: path is empty")
	}
	// start
	if path[0] != '/' {
		panic("web: not start with '/'")
	}
	// end
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: end with '/'")
	}
	// continuous in the path, be with strings.contains("//")

	// find tree first
	root, ok := r.trees[method]

	if !ok {
		// no tree
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	// handle root "/"
	if path == "/" {
		// route "/" register twice
		if root.handler != nil {
			panic("web: the route conflicts, register twice")
		}
		root.handler = handlerFunc
		return
	}

	// avoid first segment "/"
	// e.g. /user/home => divided to 3 segments
	// parse paths
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == "" {
			panic("web: no continuous '//' ")
		}
		// create node if it does not exist
		root = root.childOrCreate(seg)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: the route conflicts, %s register twice", path))
	}
	// there is a handleFunc at the leaf
	root.handler = handlerFunc
}

func (r *Router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{
			node: root,
		}, true
	}

	// Trim head and tail with "/"
	segs := strings.Split(strings.Trim(path, "/"), "/")
	var pathParams map[string]string
	for _, seg := range segs {
		child, paramChild, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		// 命中 參數路由
		if paramChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			// 參數路由格式為 :id
			pathParams[child.path[1:]] = seg
		}
		root = child
	}
	// return "true" => 不會處理node有無handler的情況
	return &matchInfo{
		node:       root,
		pathParams: pathParams,
	}, root.handler != nil
}

func (n *node) childOrCreate(seg string) *node {
	if seg == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: invalid route, there is a parameter path"))
		}
		if n.starChild == nil {
			n.starChild = &node{
				path: seg,
			}
		}
		return n.starChild
	}

	if seg[0] == ':' {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: invalid route, there is a star path"))
		}
		if n.paramChild == nil {
			n.paramChild = &node{path: seg}
		}
		return n.paramChild
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}

	res, ok := n.children[seg]
	if !ok {
		res = &node{
			path: seg,
		}
		n.children[seg] = res
	}
	return res
}

// 優先匹配靜態路由，其次通配符匹配
func (n *node) childOf(path string) (*node, bool, bool) {
	if n.children == nil {
		// 此處優先級：參數路由優先於通配符
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	res, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return res, false, ok
}
