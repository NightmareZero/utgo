package httpserv

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type httpServer struct {
	serveMux *http.ServeMux

	Ctx             context.Context
	cancel          context.CancelFunc
	Logger          Logger
	Config          ServerConfig
	ErrorHandler    RequestHandler
	NotFoundHandler RequestHandler

	middlewares map[string][]Middleware
	handleMap   map[string]urlHandler
	handle1Map  map[string]urlHandler
}

type ServerConfig struct {
	Port    int
	Timeout int64
}

func NewHttpServer(config ServerConfig) *httpServer {
	var serv = &httpServer{
		Config: config,

		middlewares: map[string][]Middleware{},
		handleMap:   make(map[string]urlHandler),
		handle1Map:  make(map[string]urlHandler),
	}
	return serv
}

func (s *httpServer) Middleware(path string, middleware Middleware) {
	_, ok := s.middlewares[path]
	if !ok {
		s.middlewares[path] = []Middleware{middleware}
	}
	s.middlewares[path] = append(s.middlewares[path], middleware)
}

func (s *httpServer) Handle(path string, method string, handler RequestHandler) {
	var hMap = s.handleMap
	if !isStaticPath(path) {
		hMap = s.handle1Map
	}

	h := hMap[path]
	if h == nil {
		h = urlHandler{}
	}
	h[strings.ToUpper(method)] = handler
	hMap[path] = h

}

func (s *httpServer) ListenAndServe() error {
	s.serveMux = http.NewServeMux()
	s.buildRouter()

	ctx := s.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(s.Config.Port),
		ReadTimeout:  time.Duration(s.Config.Timeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.Timeout) * time.Second,
		Handler:      s.serveMux,
		BaseContext: func(l net.Listener) context.Context {
			ctx2, canc := context.WithCancel(ctx)
			s.cancel = canc
			return ctx2
		},
	}
	srv.ListenAndServe()
	return nil
}

func (s *httpServer) buildRouter() {
	// TODO
}

func (s *httpServer) Stop() {
	s.cancel()
}

func isStaticPath(path string) bool {
	// TODO
	return true
}
