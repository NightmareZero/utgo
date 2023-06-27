package client_test

import (
	"testing"

	"github.com/NightmareZero/nzgoutil/client"
)

var (
	Server1   = ""
	Client1   = client.Client{Url: Server1}
	TestLogin = client.Api[map[string]any]{
		Client: &Client1,
		Path:   "/auth/login",
		Method: "POST",
		Parser: client.JsonParser,
	}
)

func TestSend(t *testing.T) {
	Client1.Init()

	a, err := TestLogin.Req(map[string]string{"username": "admin", "password": "123"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", a)

}
