package client_test

import "github.com/NightmareZero/nzgoutil/client"

var (
	Server1   = "http://localhost:8865"
	TestLogin = &client.ApiClient[any]{
		Url:       Server1,
		Path:      "/login",
		Processor: client.PostJson,
	}
)
