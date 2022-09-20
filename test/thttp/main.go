package main

import "github.com/NightmareZero/nzgoutil/hsrv"

func TestHttpDev() {
	hserv := hsrv.NewServer(hsrv.Config{
		Port: 8080,
	})
	hserv.Handle("/test", "GET", func(res hsrv.Response, req hsrv.Request) {
		s, err := req.Text()
		if err != nil {
			res.Text("err")
		}

		res.Text("hello " + s)
	})
	hserv.Middleware("/", hsrv.Middleware{})
	hserv.ListenAndServe()
}
