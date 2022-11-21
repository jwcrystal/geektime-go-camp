package web

import (
	"fmt"
	"regexp"
	"strings"
)

type nodeType int

const (
	nodeTypeStatic = iota
	nodeTypeParam
	nodeTypeAny
	nodeTypeRegexp
)

type node struct {
	// 區分路由型態
	nodeType nodeType

	// entire path
	route string

	path string
	// sub-node
	children map[string]*node

	// 通配符 "*"
	starChild *node
	//starChildren map[string]*node

	// 參數路由
	paramChild *node
	// 參數字段，與正則路由共用
	paramString string

	// 正則路由
	regChild      *node
	reqExpPattern *regexp.Regexp

	handler     HandleFunc
	middlewares []Middleware
	matchedMdls []Middleware
}

// Router tree
type Router struct {
	trees map[string]*node
}

type matchInfo struct {
	node        *node
	pathParams  map[string]string
	middlewares []Middleware
}

func NewRouter() *Router {
	return &Router{trees: map[string]*node{}}
}

// addRoute path must start with "/", not end with "/", not continues with "//", and same
// and same parameter path covered by the behind one,
// method as http method
func (r *Router) addRoute(method string, path string, handlerFunc HandleFunc, mdls ...Middleware) {
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
		root.route = "/"
		root.middlewares = mdls
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
	root.route = path
	root.middlewares = mdls
}

func (r *Router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{
			node: root,
			//middlewares: root.middlewares,
		}, true
	}

	// Trim head and tail with "/", and separate with "/"
	segs := strings.Split(strings.Trim(path, "/"), "/")
	var pathParams map[string]string
	matchInfo := &matchInfo{}
	child := root
	for _, seg := range segs {
		var paramChild, found bool
		child, paramChild, found = child.childOf(seg)
		if !found {
			// 檢查是否為通配末尾，支援多段路由
			// 可以用 type區分 ，或是 通配後字段是否結束 來區分
			// /order/*
			// /order/detail/123 (x)
			// /order/detail/123/456 (x)
			// /order/detail/123/456/789 (x)
			// 要找最後為通配的字段，所以用root，child會採用當前字段
			if root.nodeType == nodeTypeAny {
				matchInfo.node = child
				return matchInfo, true
			}
			return nil, false
		}
		// 命中 參數路由
		if paramChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			// 參數路由格式為 :id
			// 如果是正則路由，只取paraString部分
			paraString := strings.Split(child.path[1:], "(")[0]
			pathParams[paraString] = seg
		}
		//root = child
		//if len(child.middlewares) > 0 {
		//	root.middlewares = append(root.middlewares, child.middlewares...)
		//}
	}

	matchInfo.node = child
	matchInfo.pathParams = pathParams
	//matchInfo.middlewares = root.middlewares
	//matchInfo.middlewares = r.findMiddleware(root, segs)
	// return "true" => 不會處理node有無handler的情況
	return matchInfo, true
}

func (r *Router) findMiddleware(root *node, segs []string) []Middleware {
	// 遍歷匹配route的所有middlewares
	// 把 tree 整個掃過一遍，找出符合情況的middleware
	// 使用 recursive 或是 BFS、DFS
	queue := []*node{root}
	//mdlList := []Middleware{}
	mdlList := make([]Middleware, 0, 16)
	for i, _ := range segs {
		seg := segs[i]
		var children []*node
		for _, currNode := range queue {
			children = append(children, currNode.childrenOf(seg)...)
			if len(currNode.middlewares) > 0 {
				mdlList = append(mdlList, currNode.middlewares...)
			}
		}
		// 下層遍歷
		queue = children
	}
	// leaf遍歷
	for _, currNode := range queue {
		if len(currNode.middlewares) > 0 {
			mdlList = append(mdlList, currNode.middlewares...)
		}
	}
	return mdlList
}

func (r *Router) findRouteWithMiddleware(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{
			node:        root,
			middlewares: root.middlewares,
		}, true
	}

	// Trim head and tail with "/", and separate with "/"
	segs := strings.Split(strings.Trim(path, "/"), "/")
	var pathParams map[string]string
	matchInfo := &matchInfo{}
	child := root
	for _, seg := range segs {
		var paramChild, found bool
		child, paramChild, found = child.childOf(seg)
		if !found {
			// 檢查是否為通配末尾，支援多段路由
			// 可以用 type區分 ，或是 通配後字段是否結束 來區分
			// /order/*
			// /order/detail/123 (x)
			// /order/detail/123/456 (x)
			// /order/detail/123/456/789 (x)
			// 要找最後為通配的字段，所以用root，child會採用當前字段
			if root.nodeType == nodeTypeAny {
				matchInfo.node = child
				return matchInfo, true
			}
			return nil, false
		}
		// 命中 參數路由
		if paramChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			// 參數路由格式為 :id
			// 如果是正則路由，只取paraString部分
			paraString := strings.Split(child.path[1:], "(")[0]
			pathParams[paraString] = seg
		}
		//root = child
		//if len(child.middlewares) > 0 {
		//	root.middlewares = append(root.middlewares, child.middlewares...)
		//}
	}

	matchInfo.node = child
	matchInfo.pathParams = pathParams
	//matchInfo.middlewares = root.middlewares
	matchInfo.middlewares = r.findMiddleware(root, segs)
	// return "true" => 不會處理node有無handler的情況
	return matchInfo, true
}

func (n *node) childrenOf(path string) []*node {
	childNode := make([]*node, 0, 4)
	var staticChild *node
	if n.children != nil {
		staticChild = n.children[path]
	}
	// priority: any > param > static
	if n.starChild != nil {
		childNode = append(childNode, n.starChild)
	}
	if n.paramChild != nil {
		childNode = append(childNode, n.paramChild)
	}
	if staticChild != nil {
		childNode = append(childNode, staticChild)
	}
	return childNode
}

// childOrCreate to check sub-node if it does not exist
// 先判斷 path 是否為通配符路由
// 其次判斷是否為參數路由，為":"開頭
// 最後從 children 查找靜態路由
// 如果以上都沒，將會創建一個新節點 node
func (n *node) childOrCreate(seg string) *node {
	if seg == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: invalid route, a parameter path existed"))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: invalid route, a regexpr path existed"))
		}
		if n.starChild == nil {
			n.starChild = &node{path: seg, nodeType: nodeTypeAny}
		}
		return n.starChild
	}

	if seg[0] == ':' {
		paraString, regExpPattern, isRegExp := n.parseParam(seg)
		// 因原本一長串可讀性較差，重新封裝
		if isRegExp {
			return n.childOrCreateWithRegExp(seg, paraString, regExpPattern)
		}
		return n.childOrCreateWithParam(seg, paraString)
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}

	res, ok := n.children[seg]
	if !ok {
		res = &node{
			path:     seg,
			nodeType: nodeTypeStatic,
		}
		n.children[seg] = res
	}
	return res
}

// 優先匹配靜態路由，其次參數路由、通配符匹配。
// first return: child node
// second return: true if it is param or regexpr route
// third return: true if node existed
func (n *node) childOf(path string) (*node, bool, bool) {
	if n.children == nil {
		// 此處優先級：參數路由優先於通配符
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		if n.regChild != nil {
			return n.regChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	res, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		if n.regChild != nil {
			return n.regChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return res, false, ok
}

// parseParam 解析參數路由，判斷是否為正則路由
// first return: parameter path
// second return: regular expression
// third return: true as regExpr
func (n *node) parseParam(path string) (string, string, bool) {
	// path: :<param>(<regExp>)
	// remove ":"
	path = path[1:]
	segs := strings.Split(path, "(")
	//segs := strings.SplitN(path, "(",2)
	// :reg(xx) (o) => 2 segs, maybe
	// :reg(xx (x) => 2 segs, maybe
	// :reg(x(x (x) => 3 segs
	// :regxx (x)
	if len(segs) == 2 {
		// last element is ")"
		expr := segs[1]
		if expr[len(expr)-1] == ')' {
			return segs[0], expr[:len(expr)-1], true
		}
	}
	return path, "", false
}

func (n *node) childOrCreateWithParam(path string, paraString string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: invalid route, a star path existed"))
	}
	if n.paramChild != nil {
		if n.paramChild.path != path {
			panic(fmt.Sprintf("web: parameter route conflict, had %s, new %s", n.paramChild.path, path))
		}
	}
	if n.regChild != nil {
		panic(fmt.Sprintf("web: regexpr route conflict, had %s, new %s", n.regChild.path, path))
	}

	n.paramChild = &node{path: path, paramString: paraString, nodeType: nodeTypeParam}
	return n.paramChild
}

func (n *node) childOrCreateWithRegExp(path string, paraString string, regExpPattern string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: invalid route, a star path existed"))
	}
	if n.paramChild != nil {
		if n.paramChild.path != path {
			panic(fmt.Sprintf("web: parameter route conflict, had %s, new %s", n.paramChild.path, path))
		}
	}
	if n.regChild != nil {
		if n.regChild.reqExpPattern.String() != regExpPattern || n.paramString != paraString {
			panic(fmt.Sprintf("web: regexpr route conflict, had %s, new %s", n.regChild.path, path))
		}
	}
	regExp, err := regexp.Compile(regExpPattern)
	if err != nil {
		panic(fmt.Errorf("web: regexpr error: %w", err))
	}
	n.regChild = &node{path: path, paramString: paraString, reqExpPattern: regExp, nodeType: nodeTypeRegexp}
	return n.regChild
}
