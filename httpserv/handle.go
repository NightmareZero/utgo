package httpserv

type Middleware struct {
	BeforeRequest func(Request, Response) bool
	AfterRequest  func(Request, Response)
}

type urlHandler map[string]RequestHandler

func (u urlHandler) ServeRequest(server *httpServer, request Request, response Response) {
	defer func() {
		i := recover()
		if i != nil {
			if server.ErrorHandler != nil {
				func(request Request, response Response) {
					defer func() {
						i2 := recover()
						if i2 != nil {
							if server.Logger != nil {
								server.Logger.Errorf("request error, url: %v,%+v", request.URL, i2)
							}
						}
					}()
					server.ErrorHandler(request, response)
				}(request, response)
			}
		}
	}()

	rh := u[request.Method]
	if rh == nil {
		if server.NotFoundHandler != nil {
			server.NotFoundHandler(request, response)
		}
		return
	}
	rh(request, response)
}
