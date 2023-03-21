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
	"sync"
	"time"

	"github.com/rs/cors"
)

type CtxTag string

const WebContextVName CtxTag = "WebContext"

type Server struct {
	serveMux *http.ServeMux
	rl       *sync.Mutex // 运行锁

	Ctx             context.Context         // 全局上下文
	cancel          context.CancelFunc      // 终止函数
	Logger          Logger                  // 日志输出
	Config          Config                  // 配置
	ErrorHandler    ErrorHandler            // 统一错误处理
	NotFoundHandler RequestHandler          // 统一404处理
	CtxDataGetter   func(*http.Request) any // 上下文生成器

	middleware []_middleware         // 内部 中间件列表
	handleMap  map[string]urlHandler // 内部 路由表

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
		Config:          config,
		rl:              &sync.Mutex{},
		Logger:          defaultLogger,
		handleMap:       map[string]urlHandler{},
		ErrorHandler:    defaultPanicHandler,
		NotFoundHandler: defaultNotFoundHandler,
		CtxDataGetter:   func(r *http.Request) any { return nil },
	}
	return serv
}

// function around handler
// prefix: url prefix for interceptor
func (s *Server) Interceptor(prefix string, interceptor Interceptor) {
	s.rl.Lock()
	defer s.rl.Unlock()

	s.middleware = append(s.middleware, _middleware{
		prefix: prefix,
		before: interceptor,
	})
	sort.SliceStable(s.middleware, func(i, j int) bool {
		return s.middleware[i].before.Order() < s.middleware[j].before.Order()
	})
}

func (s *Server) PostProcessor(prefix string, postProcessor PostProcessor) {
	s.rl.Lock()
	defer s.rl.Unlock()

	s.middleware = append(s.middleware, _middleware{
		prefix: prefix,
		after:  postProcessor,
	})
	sort.SliceStable(s.middleware, func(i, j int) bool {
		return s.middleware[i].before.Order() < s.middleware[j].before.Order()
	})
}

// listen method on path
// path: listen path (if end with '/' ,will listen all start with $path)
// method: listen method such as 'GET', 'POST', 'PUT', 'DELETE'
// handler: function handler
// =======
// 监听方法
// path: 路径 (如果以 '/' 为结尾，则会监听所有以$path开头的路径)
// method: 监听的服务器方法
// handler: 执行方法的句柄
func (s *Server) Handle(path string, method string, handler RequestHandler) {
	s.rl.Lock()
	defer s.rl.Unlock()

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

func (s *Server) Static(path, static string) {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	s.Handle(path, "GET", newStaticFileHandler(s, path, static))
}

func (s *Server) ListenAndServe() error {
	s.Logger.Info("hSrv: server starting...")
	s.rl.Lock()
	defer s.rl.Unlock()

	s.serveMux = http.NewServeMux()
	s.buildRouter()

	if s.Ctx == nil {
		s.Ctx = context.Background()
	}
	ctx := context.WithValue(s.Ctx, WebContextVName, s.WebContext)

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
		s.Logger.Info("hSrv: tls enabled")
		// 初始化x509 certificate
		certPool := x509.CertPool{}
		s.Logger.Info("hSrv: load ca : %v", s.Config.Tls.CaPath)
		b, err := os.ReadFile(s.Config.Tls.CaPath)
		if err != nil {
			return fmt.Errorf("hSrv: failed to read ca files, %w", err)
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
		s.Logger.Infof("hSrv: serve https, port %v", s.Config.Port)
		return srv.ListenAndServeTLS(s.Config.Tls.CrtPath, s.Config.Tls.KeyPath)
	}

	s.Logger.Infof("hSrv: serve http, port %v", s.Config.Port)
	return srv.ListenAndServe()
}

func (s *Server) Stop() {
	s.cancel()
}
