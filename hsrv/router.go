package hsrv

import (
	"net/http"
	"strings"
)

func (s *Server) buildRouter() {

	for url, urlHandler := range s.handleMap {
		s.Logger.Debugf("hSrv: listen %v", url)
		var mdi []Interceptor
		var mdp []PostProcessor
		for k2, mas := range s.middleware {
			if strings.HasPrefix(url, mas.prefix) {
				if mas.before != nil {
					mdi = append(mdi, s.middleware[k2].before)
				}
				if mas.after != nil {
					mdp = append(mdp, s.middleware[k2].after)
				}
			}
		}

		if len(mdi) > 0 || len(mdp) > 0 {
			router := middlewareRouter{
				u:      urlHandler,
				before: mdi,
				after:  mdp,
			}
			s.serveMux.Handle(url, router)
		} else {
			s.serveMux.Handle(url, urlHandler)
		}

	}
}

type _middleware struct {
	prefix string // 拦截路径
	before Interceptor
	after  PostProcessor
}

type Interceptor interface {
	Order() int // 顺序
	Before(Response, Request) bool
}

type PostProcessor interface {
	Order() int // 顺序
	After(Response) bool
}

type middlewareRouter struct {
	u      urlHandler
	before []Interceptor
	after  []PostProcessor
}

type reqCtx struct {
	Data   any
	after  []PostProcessor
	Server *Server
	Req    *http.Request
	Res    http.ResponseWriter
}

func (u middlewareRouter) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	reqCtx := reqCtx{
		Data:   u.u.s.CtxDataGetter(request),
		Server: u.u.s,
		after:  u.after,
		Req:    request,
		Res:    response,
	}

	req := Request{request, &reqCtx}
	res := Response{response, &reqCtx}
	defer requestRecover(u.u.s, res, req)

	for _, m := range u.before {
		if !m.Before(res, req) {
			return
		}
	}

	u.u.serveHTTP(res, req)

	// for _, m := range u.mds {
	// 	m.After(res, req)
	// }
}
