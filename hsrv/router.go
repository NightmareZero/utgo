package hsrv

import (
	"net/http"
	"strings"
)

func (s *hsrver) buildRouter() {
	var md []Middleware = make([]Middleware, len(s.middlewares))
	for k, uh := range s.handleMap {
		mdCache := md[:0]
		for _, mas := range s.middlewares {
			if strings.HasPrefix(k, mas.prefix) {
				mdCache = append(md, mas)
			}
		}

		if len(mdCache) > 0 {
			router := middlewaredRouter{
				u: uh,
			}
			router.mds = append(router.mds, mdCache...)
			s.serveMux.Handle(k, router)
		} else {
			s.serveMux.Handle(k, uh)
		}

	}
	// TODO
}

type Middleware struct {
	prefix string // 拦截路径
	Order  int    // 顺序
	Before func(Response, Request) bool
	After  func(Response, Request)
}

type middlewaredRouter struct {
	u   urlHandler
	mds []Middleware
}

func (u middlewaredRouter) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	req, res := Request{request}, Response{response}
	defer defaultRecover(u.u.s, res, req)

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
