package httpserv

import "net/http"

type Middleware struct {
	Before func(Request, Response) bool
	After  func(Request, Response)
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
		default405Handler(request, response)
		return
	}
	rh(request, response)
}

func default405Handler(request Request, response Response) {
	response.Status(http.StatusMethodNotAllowed)
	response.Text("method not allowed")
}
