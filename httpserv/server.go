package httpserv

import (
	"context"
	"net"
	"net/http"
	"sort"
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

	middlewares []Middleware
	handleMap   map[string]urlHandler
}

type ServerConfig struct {
	Port    int
	Timeout int64
}

func NewHttpServer(config ServerConfig) *httpServer {
	var serv = &httpServer{
		Config: config,

		handleMap: make(map[string]urlHandler),
	}
	return serv
}

func (s *httpServer) Middleware(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
	sort.SliceStable(s.middlewares, func(i, j int) bool {
		return s.middlewares[i].Prefix < s.middlewares[j].Prefix
	})
}

func (s *httpServer) Handle(path string, method string, handler RequestHandler) {
	h := s.handleMap[path]
	if h.router == nil {
		h = urlHandler{}
	}
	h.router[strings.ToUpper(method)] = handler
	s.handleMap[path] = h

}

func (s *httpServer) ListenAndServe() error {
	s.serveMux = http.NewServeMux()
	if s.ErrorHandler == nil {
		s.ErrorHandler = defaultPanicHandler
	}
	if s.NotFoundHandler == nil {
		s.NotFoundHandler = defaultNotFoundHandler
	}
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

func (s *httpServer) Stop() {
	s.cancel()
}
