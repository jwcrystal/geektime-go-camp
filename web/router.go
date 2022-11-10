package web

import "strings"

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

func (r *Router) AddRoute(method string, path string, handlerFunc HandleFunc) {
	// find tree
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
		root.handler = handlerFunc
		return
	}

	// avoid first segment "/"
	// e.g. /user/home => divided to 3 segments
	path = path[1:]
	// parse paths
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		// create node if it does not exist
		children := root.childOrCreate(seg)
		root = children
	}
	// there is a handleFunc at the leaf
	root.handler = handlerFunc
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
