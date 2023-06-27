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
	reqCtx := reqCtx{
		Data:   map[string]any{},
		Server: u.s,
		after:  nil,
	}
	ctx := Ctx{response, request, &reqCtx}
	u.serveHTTP(ctx)
}

func (u urlHandler) serveHTTP(r Ctx) {
	defer requestRecover(u.s, r)

	rh := u.router[r.R.Method]
	if rh == nil {
		defaultNotFoundHandler(r)
		return
	}
	rh(r)
}

func requestRecover(s *Server, ctx Ctx) {
	i := recover()
	if i != nil {
		if s.ErrorHandler != nil {
			doRecover(i, s, ctx)
		}
	}
}

func doRecover(i any, s *Server, ctx Ctx) {
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

	func(r Ctx) {
		defer func() {
			i2 := recover()
			if i2 != nil {
				if s.Logger != nil {
					s.Logger.Errorf("request error, url: %v,%+v", r.R.URL, i2)
				}
			}
		}()
		s.ErrorHandler(r, err)
	}(ctx)
}

func defaultNotFoundHandler(r Ctx) {
	r.Text(http.StatusNotFound, "path not found")

}

func defaultPanicHandler(r Ctx, err error) {
	r.Server.Logger.Errorf("%+v", err)

	r.Text(http.StatusInternalServerError, "internal server error")
}

func newStaticFileHandler(s *Server, basepath, staticpath string) RequestHandler {
	return func(r Ctx) {
		p := r.R.URL.Path
		truepath := staticpath + strings.Replace(p, basepath, "", 1)

		f, err := os.OpenFile(truepath, os.O_RDWR, 0)
		if err != nil {
			s.Logger.Errorf("hSrv: file not found url: %v static: %v", p, truepath)
			r.Text(http.StatusNotFound, "path not found")
			return
		}
		defer f.Close()

		fi, _ := f.Stat()

		r.File(f, fi.Size(), f.Name())
	}
}
