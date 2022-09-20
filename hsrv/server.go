package hsrv

import (
	"context"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type hsrver struct {
	serveMux *http.ServeMux

	Ctx             context.Context
	cancel          context.CancelFunc
	Logger          Logger
	Config          Config
	ErrorHandler    RequestHandler
	NotFoundHandler RequestHandler

	middlewares []Middleware
	handleMap   map[string]urlHandler
}

type Config struct {
	Port    int
	Timeout int64
}

func NewServer(config Config) *hsrver {
	var serv = &hsrver{
		Config:    config,
		Logger:    defaultLogger,
		handleMap: make(map[string]urlHandler),
	}
	return serv
}

func (s *hsrver) Middleware(prefix string, middleware Middleware) {
	middleware.prefix = prefix
	s.middlewares = append(s.middlewares, middleware)
	sort.SliceStable(s.middlewares, func(i, j int) bool {
		return s.middlewares[i].prefix < s.middlewares[j].prefix
	})
}

func (s *hsrver) Handle(path string, method string, handler RequestHandler) {
	h := s.handleMap[path]
	if h.router == nil {
		h = urlHandler{
			router: map[string]RequestHandler{},
		}
	}
	h.router[strings.ToUpper(method)] = handler
	s.handleMap[path] = h

}

func (s *hsrver) ListenAndServe() error {
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

func (s *hsrver) Stop() {
	s.cancel()
}
