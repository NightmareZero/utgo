package hsrv

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type urlHandler struct {
	s      *Server
	router map[string]RequestHandler
}

func (u urlHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	u.serveHTTP(Response{response}, Request{request, u.s, u.s.requestCtxGetter(request)})
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
	stack := make([]byte, 4096)

	size := 0
	for {
		size = runtime.Stack(stack, false)
		// The size of the buffer may be not enough to hold the stacktrace,
		// so double the buffer size
		if size == len(stack) {
			stack = make([]byte, len(stack)<<1)
			continue
		}
		break
	}
	switch ii := i.(type) {
	case error:
		err = fmt.Errorf("panic: %+v,%v", ii, string(stack[:size]))
	case string:
		err = fmt.Errorf("panic: %+v,%v", ii, string(stack[:size]))
	case int:
		err = fmt.Errorf("error code: %v,%v", ii, string(stack[:size]))
	default:
		err = fmt.Errorf("panic: %v", string(stack[:size]))
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

func defaultNotFoundHandler(w Response, r Request) {
	w.Text("path not found", http.StatusNotFound)

}

func defaultPanicHandler(w Response, r Request, err error) {
	r.Server.Logger.Errorf("%+v", err)

	w.Text("internal server error", http.StatusInternalServerError)
}

func newStaticFileHandler(s *Server, basepath, staticpath string) RequestHandler {
	return func(w Response, r Request) {
		p := r.URL.Path
		truepath := staticpath + strings.Replace(p, basepath, "", 1)

		f, err := os.OpenFile(truepath, os.O_RDWR, 0)
		if err != nil {
			s.Logger.Errorf("hSrv: file not found url: %v static: %v", p, truepath)
			w.Text("path not found", http.StatusNotFound)
			return
		}
		defer f.Close()

		fi, _ := f.Stat()

		w.File(f, fi.Size(), f.Name())
	}
}
