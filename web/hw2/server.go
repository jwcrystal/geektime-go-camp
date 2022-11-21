package web

import (
	"log"
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

type Server interface {
	http.Handler
	Start(addr string) error
	addRoute(method string, path string, handler HandleFunc, mdls ...Middleware)
}

// tips, 確保 HttpServer 確實實現了 Server 接口
var _ Server = &HttpServer{}

type HttpServer struct {
	router      *Router
	middlewares []Middleware
}
type HttpServerOption func(server *HttpServer)

func NewHttpServer() *HttpServer {
	return &HttpServer{
		NewRouter(),
		nil,
	}
}

// NewHttpServerV1 擴展性不好
func NewHttpServerV1(middlewares ...Middleware) *HttpServer {
	return &HttpServer{
		router:      NewRouter(),
		middlewares: middlewares,
	}
}

// NewHttpServerV2 Option模式
func NewHttpServerV2(opts ...HttpServerOption) *HttpServer {
	res := &HttpServer{
		router: NewRouter(),
	}

	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithMiddlewares(mdls ...Middleware) HttpServerOption {
	return func(server *HttpServer) {
		server.middlewares = mdls
	}
}

// ServerHTTP is the entry endpoint of HttpServer
func (h *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// endpoint logic
	ctx := &Context{
		Req: request,
		Res: writer,
	}
	// find the routes, and launch handleFunc
	//h.serve(ctx)
	// server級別的中間件(middleware)
	root := h.serve // last one
	// bind the middlewares
	// 從後到前組裝
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		//h.middlewares = append(h.middlewares)
		root = h.middlewares[i](root)
	}
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (s *HttpServer) Use(mdls ...Middleware) {
	if s.middlewares == nil {
		s.middlewares = mdls
		return
	}
	s.middlewares = append(s.middlewares, mdls...)
}

func (h *HttpServer) UseV1(method string, path string, mdls ...Middleware) {
	h.addRoute(method, path, nil, mdls...)
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

func (h *HttpServer) addRoute(method string, path string, handler HandleFunc, mdls ...Middleware) {
	h.router.addRoute(method, path, handler, mdls...)
}

func (h *HttpServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HttpServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HttpServer) serve(ctx *Context) {
	// find route
	// before route
	route, ok := h.router.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	// after route
	if !ok || route.node.handler == nil || route.node == nil {
		ctx.Res.WriteHeader(http.StatusNotFound)
		ctx.Res.Write([]byte("Not found"))
		return
	}
	ctx.PathParams = route.pathParams
	if route.node != nil {
		ctx.MatchedRoute = route.node.route
	}
	// before execute
	route.node.handler(ctx)
	// after execute
}

func (s *HttpServer) flashResp(ctx *Context) {
	if ctx.ResStatusCode > 0 {
		ctx.Res.WriteHeader(ctx.ResStatusCode)
	}
	_, err := ctx.Res.Write(ctx.ResData)
	if err != nil {
		log.Fatalln("回写响应失败", err)
	}
}
