package hsrv

import (
	"fmt"
	"net/http"
	"runtime"
)

type urlHandler struct {
	s      *Server
	router map[string]RequestHandler
}

func (u urlHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	u.serveHTTP(Response{response}, Request{request, u.s.WebContext})
}

func (u urlHandler) serveHTTP(response Response, request Request) {
	defer requestRecover(u.s, response, request)

	rh := u.router[request.Method]
	if rh == nil {
		defaultNotFoundHandler(response, request)
		return
	}
	rh(response, request)
}

func requestRecover(s *Server, response Response, request Request) {
	i := recover()
	if i != nil {
		if s.ErrorHandler != nil {
			doRecover(i, s, request, response)
		}
	}
}

func doRecover(i any, s *Server, request Request, response Response) {
	var err error
	var stack [4096]byte
	runtime.Stack(stack[:], false)
	switch ii := i.(type) {
	case error:
		err = fmt.Errorf("panic: %+v,%+v", ii, string(stack[:]))
	case string:
		err = fmt.Errorf("panic: %+v,%+v", ii, string(stack[:]))
	case int:
		err = fmt.Errorf("error code: %v,%+v", ii, string(stack[:]))
	default:
		err = fmt.Errorf("panic: %+v", string(stack[:]))
	}

	func(request Request, response Response) {
		defer func() {
			i2 := recover()
			if i2 != nil {
				if s.Logger != nil {
					s.Logger.Errorf("request error, url: %v,%+v", request.URL, i2)
				}
			}
		}()
		s.ErrorHandler(response, request, err)
	}(request, response)
}

func defaultNotFoundHandler(response Response, request Request) {
	response.Text("path not found", http.StatusNotFound)

}

func defaultPanicHandler(response Response, request Request, err error) {
	defaultLogger.Errorf("%+v", err)

	response.Text("internal server error", http.StatusInternalServerError)
}
