package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// 常规请求定义
type Api[R any] struct {
	Method string // 默认视为POST
	Parser DataParser
	Path   string
	Client *Client
}

type Client struct {
	Url     string
	CrtPath string
	KeyPath string
	Timeout int

	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConns            int
	*http.Client
}

// 执行网络请求
//
// req: 请求方 body,
// res: 接收方 body(要求为指针类型),
// h:   需要的 http-header,
// qp:  url中的参数，要求类似 "id=123" 样式,
func (a *Api[R]) Req(req any, h map[string]string, qp ...string) (R, error) {
	var r R
	return r, a.doReq(context.Background(), req, &r, h, qp...)
}

func (a *Api[R]) ReqWithCtx(ctx context.Context, req any, h map[string]string, qp ...string) (R, error) {
	var r R
	return r, a.doReq(ctx, req, &r, h, qp...)
}

// 执行请求
func (def *Api[R]) doReq(ctx context.Context, req any, res *R, headers map[string]string, qp ...string) (err error) {
	if def.Client == nil {
		return fmt.Errorf("requester.req client is nil")
	}

	// 获取url
	url := def.Path

	// 如果有其他参数
	if len(qp) > 0 {
		url += "?" + strings.Join(qp, ";")
	}

	// 获取请求body
	b := []byte{}
	if req != nil && def.Parser.R != nil {
		b, err = def.Parser.R(req)
		if err != nil {
			return fmt.Errorf("requester.req get request body bytes: %w", err)
		}
	}

	// 生成请求
	request, err := http.NewRequest(def.Method, def.Client.Url+url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("requester.req new request: %w", err)
	}
	if nil != ctx {
		request = request.WithContext(ctx)
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}

	//处理返回结果
	response, err := def.Client.Do(request)
	if err != nil {
		return fmt.Errorf("requester.req new request: %+v", err)
	}
	defer response.Body.Close()

	b, err = io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("requester.req read body error: %+v", err)
	}

	err = def.Parser.W(b, res)
	if err != nil {
		return fmt.Errorf("requester.req convert body %v, error: %+v", string(b), err)
	}

	return nil

}

func (cl *Client) Init() error {
	cl.Default()
	tr := &http.Transport{
		MaxIdleConns:        cl.MaxIdleConns,
		MaxIdleConnsPerHost: cl.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cl.MaxConns,
		DialContext:         (&net.Dialer{Timeout: time.Duration(cl.Timeout) * time.Second}).DialContext,
	}
	c := &http.Client{Transport: tr}

	if len(cl.CrtPath) > 0 {
		pool := x509.NewCertPool()
		// 这里加载服务端提供的证书，用于校验服务端返回的数据
		aCrt, err := os.ReadFile(cl.CrtPath)
		if err != nil {
			return err
		}
		pool.AppendCertsFromPEM(aCrt)
		// 这里加载客户端自己的证书，要与提供给服务端的证书一致，不然服务端校验会不通过
		cliCrt, err := tls.LoadX509KeyPair(cl.CrtPath, cl.KeyPath)
		if err != nil {
			return err
		}
		tr.TLSClientConfig = &tls.Config{
			RootCAs:            pool,
			Certificates:       []tls.Certificate{cliCrt},
			InsecureSkipVerify: true,
		}
	}

	cl.Client = c

	return nil
}

func (cl *Client) Default() {
	if cl.MaxConns == 0 {
		cl.MaxConns = 30
	}
	if cl.MaxIdleConns == 0 {
		cl.MaxIdleConns = 10
	}
	if cl.MaxIdleConnsPerHost == 0 {
		cl.MaxIdleConnsPerHost = 10
	}
	if cl.Timeout == 0 {
		cl.Timeout = 10
	}
	cl.Url = strings.TrimRight(cl.Url, "/")
}

type DataParser struct {
	R func(req any) ([]byte, error) // RequestReader
	W func(b []byte, res any) error // ResponseWriter
}

var (
	NoParser       = DataParser{R: ByteReader, W: ByteWriter}
	JsonParser     = DataParser{R: JsonReader, W: JsonWriter}
	TextParser     = DataParser{R: TextReader, W: TextWriter}
	FormParser     = DataParser{} // 暂时不实现
	GetFileParser  = DataParser{} // 暂时不实现
	SendFileParser = DataParser{} // 暂时不实现
)

func JsonReader(req any) ([]byte, error) {
	return json.Marshal(req)
}

func JsonWriter(b []byte, res any) error {
	return json.Unmarshal(b, res)
}

func TextReader(req any) ([]byte, error) {
	if b, ok := req.([]byte); ok {
		return b, nil
	}
	return nil, fmt.Errorf("invalid input type, need []byte")
}

func TextWriter(b []byte, res any) error {
	if rb, ok := res.([]byte); ok {
		rb = append(rb, b...)
		res = rb
		return nil
	}
	return fmt.Errorf("invalid putput type, need []byte")
}

func ByteReader(req any) ([]byte, error) {
	return req.([]byte), nil
}

func ByteWriter(b []byte, res any) error {
	res = b
	return nil
}
