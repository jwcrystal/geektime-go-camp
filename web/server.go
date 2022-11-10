package web

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

type Server interface {
	http.Handler
	Start(addr string) error
	addRoute(method string, path string, handler HandleFunc)
}

// tips, 確保 HttpServer 確實實現了 Server 接口
var _ Server = &HttpServer{}

type HttpServer struct {
	*Router
}

func (h *HttpServer) NewHttpServer() *HttpServer {
	return &HttpServer{
		NewRouter(),
	}
}

// ServerHTTP is the entry endpoint of HttpServer
func (h *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// endpoint logic
	ctx := &Context{
		req: request,
		res: writer,
	}
	// find the routes, and launch handleFunc
	h.serve(ctx)
}

func (h *HttpServer) Start(addr string) error {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// here, do "after start" callback
	// start with what you need (前置條件)

	return http.Serve(l, h)
}

func (h *HttpServer) Start1(addr string) error {
	return http.ListenAndServe(addr, h)
}

func (h *HttpServer) addRoute(method string, path string, handler HandleFunc) {

}

func (h *HttpServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HttpServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HttpServer) serve(ctx *Context) {
	// find route
	//route, ok := h.findRoute(ctx.req.Method, ctx.req.URL.Path)
	//if !ok || route.trees{
	//
	//}
}
