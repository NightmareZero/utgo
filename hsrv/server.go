package hsrv

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
)

type Server struct {
	serveMux *http.ServeMux

	Ctx             context.Context    // 全局上下文
	cancel          context.CancelFunc // 终止函数
	Logger          Logger             // 日志输出
	Config          Config             // 配置
	ErrorHandler    ErrorHandler       // 统一错误处理
	NotFoundHandler RequestHandler     // 统一404处理

	middlewares []_middleware         // 内部 中间件列表
	handleMap   map[string]urlHandler // 内部 路由表

	WebContext any // 全局web上下文(所有连接共享), 会被放到Request中
}

type Config struct {
	Port    int
	Timeout int64
	Tls     *TlsConfig
}

type TlsConfig struct {
	CrtPath string
	KeyPath string
	CaPath  string
	Cors    []string
}

func NewServer(config Config) *Server {
	var serv = &Server{
		Config:    config,
		Logger:    defaultLogger,
		handleMap: make(map[string]urlHandler),
	}
	return serv
}

func (s *Server) Middleware(prefix string, middleware Middleware) {
	s.middlewares = append(s.middlewares, _middleware{
		prefix: prefix,
		md:     middleware,
	})
	sort.SliceStable(s.middlewares, func(i, j int) bool {
		return s.middlewares[i].md.Order() < s.middlewares[j].md.Order()
	})
}

func (s *Server) Handle(path string, method string, handler RequestHandler) {
	h := s.handleMap[path]
	if h.router == nil {
		h = urlHandler{
			s:      s,
			router: map[string]RequestHandler{},
		}
	}
	h.router[strings.ToUpper(method)] = handler
	s.handleMap[path] = h

}

func (s *Server) ListenAndServe() error {
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

	// 如果添加了证书路径
	if s.Config.Tls != nil {
		// 初始化x509 certificate
		certPool := x509.CertPool{}
		b, err := os.ReadFile(s.Config.Tls.CaPath)
		if err != nil {
			return fmt.Errorf("failed to read ca files, %w", err)
		}
		certPool.AppendCertsFromPEM(b)

		srv.TLSConfig = &tls.Config{
			ClientCAs:  &certPool,
			ClientAuth: tls.RequireAnyClientCert,
		}

		// 跨域配置
		if len(s.Config.Tls.Cors) > 0 {
			c := cors.New(cors.Options{
				AllowedOrigins: s.Config.Tls.Cors,
				AllowedMethods: []string{http.MethodPost, http.MethodGet, http.MethodPut,
					http.MethodPatch, http.MethodDelete, http.MethodOptions},
				MaxAge:         600,
				AllowedHeaders: []string{"*"},
			})
			srv.Handler = c.Handler(s.serveMux)
		}
		return srv.ListenAndServeTLS(s.Config.Tls.CrtPath, s.Config.Tls.KeyPath)
	}

	return srv.ListenAndServe()
}

func (s *Server) Stop() {
	s.cancel()
}
