package hsrv

import (
	"net/http"
	"strings"
)

func (s *Server) buildRouter() {

	for k, uh := range s.handleMap {
		s.Logger.Infof("hSrv: listen %v", k)
		var md []Middleware
		for k2, mas := range s.middlewares {
			if strings.HasPrefix(k, mas.prefix) {
				md = append(md, s.middlewares[k2].md)
			}
		}

		if len(md) > 0 {
			router := middlewaredRouter{
				u:   uh,
				mds: md,
			}
			s.serveMux.Handle(k, router)
		} else {
			s.serveMux.Handle(k, uh)
		}

	}
}

type _middleware struct {
	prefix string // 拦截路径
	md     Middleware
}

type Middleware interface {
	Order() int // 顺序
	Before(Response, Request) bool
	After(Response, Request)
}

type middlewaredRouter struct {
	u   urlHandler
	mds []Middleware
}

func (u middlewaredRouter) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	req, res := Request{request}, Response{response}
	defer requestRecover(u.u.s, res, req)

	for _, m := range u.mds {
		if !m.Before(res, req) {
			return
		}
	}

	u.u.serveHTTP(res, req)

	for _, m := range u.mds {
		m.After(res, req)
	}
}
