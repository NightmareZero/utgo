package httpserv

import "net/http"

type urlHandler struct {
	s      *httpServer
	router map[string]RequestHandler
}

func (u urlHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	u.serveHTTP(Response{response}, Request{request})
}

func (u urlHandler) serveHTTP(response Response, request Request) {
	defer defaultRecover(u.s, response, request)

	rh := u.router[request.Method]
	if rh == nil {
		defaultNotFoundHandler(response, request)
		return
	}
	rh(response, request)
}

func defaultRecover(s *httpServer, response Response, request Request) {
	i := recover()
	if i != nil {
		if s.ErrorHandler != nil {
			func(request Request, response Response) {
				defer func() {
					i2 := recover()
					if i2 != nil {
						if s.Logger != nil {
							s.Logger.Errorf("request error, url: %v,%+v", request.URL, i2)
						}
					}
				}()
				s.ErrorHandler(response, request)
			}(request, response)
		}
	}
}

func defaultNotFoundHandler(response Response, request Request) {
	response.Status(http.StatusNotFound)
	response.Text("path not found")
}

func defaultPanicHandler(response Response, request Request) {
	response.Status(http.StatusInternalServerError)
	response.Text("internal server error")
}
