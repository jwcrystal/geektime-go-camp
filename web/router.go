package web

import (
	"fmt"
	"strings"
)

type node struct {
	path string
	// sub-node
	children map[string]*node
	// non-core operation
	handler HandleFunc
}

// Router tree
type Router struct {
	trees map[string]*node
}

func NewRouter() *Router {
	return &Router{trees: map[string]*node{}}
}

// Add the restriction

// addRoute path must start with "/", not end with "/", not continues with "//"
func (r *Router) addRoute(method string, path string, handlerFunc HandleFunc) {
	if path == "" {
		panic("web: path is empty")
	}
	// find tree first
	root, ok := r.trees[method]

	if !ok {
		// no tree
		root = &node{
			path: "/",
		}
		r.trees[method] = root
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
		child := root.childOrCreate(seg)
		root = child
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: the route conflicts, %s register twice", path))
	}
	// there is a handleFunc at the leaf
	root.handler = handlerFunc
}

func (r *Router) findRoute(method string, path string) (*node, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return root, true
	}

	// Trim head and tail with "/"
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segs {
		child, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		root = child
	}
	// return "true" => 不會處理node有無handler的情況
	//return root, true
	return root, root.handler != nil
}

func (n *node) childOrCreate(seg string) *node {
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

func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return nil, false
	}
	res, ok := n.children[path]
	return res, ok
}
