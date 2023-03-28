package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// 常规请求定义
type ApiClient[R any] struct {
	Url       string // 随机选一个
	Processor ApiProcessor
	Path      string
	Conf      Conf
	cl_inited sync.Once
	client    *http.Client
}

type Conf struct {
	CrtPath string
	KeyPath string
	Timeout int
}

// 执行网络请求
//
// req: 请求方 body,
// res: 接收方 body(要求为指针类型),
// h:   需要的 http-header,
// qp:  url中的参数，要求类似 "id=123" 样式,
func (a ApiClient[R]) Req(req any, h map[string]string, qp ...string) (R, error) {
	var r R
	return r, a.doReq(context.Background(), req, &r, h, qp...)
}

func (a ApiClient[R]) ReqWithCtx(ctx context.Context, req any, h map[string]string, qp ...string) (R, error) {
	var r R
	return r, a.doReq(ctx, req, &r, h, qp...)
}

// 执行请求
func (def ApiClient[R]) doReq(ctx context.Context, req any, res *R, headers map[string]string, qp ...string) (err error) {
	// 获取url
	url := strings.TrimRight(def.Url, "/")
	url += def.Path

	// 如果有其他参数
	if len(qp) > 0 {
		url += "?" + strings.Join(qp, ";")
	}

	def.cl_inited.Do(func() {
		// 获取client
		def.client, err = getClient(def.Conf)
	})
	if err != nil {
		return fmt.Errorf("requester.req getClient: %w", err)
	}

	// 获取请求body
	b := []byte{}
	if req != nil && def.Processor.r != nil {
		b, err = def.Processor.r(req)
		if err != nil {
			return fmt.Errorf("requester.req get request body bytes: %w", err)
		}
	}

	// 生成请求
	request, err := http.NewRequest(def.Processor.method, url, bytes.NewReader(b))
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
	response, err := def.client.Do(request)
	if err != nil {
		return fmt.Errorf("requester.req new request: %+v", err)
	}
	defer response.Body.Close()

	b, err = io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("requester.req read body error: %+v", err)
	}

	err = def.Processor.w(b, res)
	if err != nil {
		return fmt.Errorf("requester.req convert body error: %+v", err)
	}

	return nil

}

func getClient(conf Conf) (*http.Client, error) {
	tr := &http.Transport{}
	c := &http.Client{Transport: tr}

	if len(conf.CrtPath) > 0 {
		pool := x509.NewCertPool()
		// 这里加载服务端提供的证书，用于校验服务端返回的数据
		aCrt, err := os.ReadFile(conf.CrtPath)
		if err != nil {
			fmt.Println("requester.getClient err, ", err)
			return nil, err
		}
		pool.AppendCertsFromPEM(aCrt)
		// 这里加载客户端自己的证书，要与提供给服务端的证书一致，不然服务端校验会不通过
		cliCrt, err := tls.LoadX509KeyPair(conf.CrtPath, conf.KeyPath)
		if err != nil {
			fmt.Println("requester.getClient: Loadx509keypair, ", err)
			return nil, err
		}
		tr.TLSClientConfig = &tls.Config{
			RootCAs:            pool,
			Certificates:       []tls.Certificate{cliCrt},
			InsecureSkipVerify: true,
		}
	}

	return c, nil
}

type ApiProcessor struct {
	method string
	r      func(req any) ([]byte, error) // RequestReader
	w      func(b []byte, res any) error // ResponseWriter
}

var (
	GetJson    = ApiProcessor{method: "GET", w: jsonw}            // 返回类型: json -> *struct
	PostJson   = ApiProcessor{method: "POST", r: jsonr, w: jsonw} // 请求类型 struct -> json  返回类型 json -> *struct
	PutJson    = ApiProcessor{method: "PUT", r: jsonr, w: jsonw}  // 请求类型 struct -> json  返回类型 json -> *struct
	DelJson    = ApiProcessor{method: "DELETE", w: jsonw}         // 返回类型: json -> *struct
	GetText    = ApiProcessor{method: "GET", w: txtw}             // 返回类型 []byte
	PostText   = ApiProcessor{method: "POST", r: txtr, w: txtw}   // 请求类型 []byte  返回类型 []byte
	PutText    = ApiProcessor{method: "PUT", r: txtr, w: txtw}    // 请求类型 []byte  返回类型 []byte
	DeleteText = ApiProcessor{method: "DELETE", w: txtw}          // 返回类型 []byte
	PostForm   = ApiProcessor{method: "POST"}                     // 暂时不实现
	PostFile   = ApiProcessor{method: "POST"}                     // 暂时不实现
)

func jsonr(req any) ([]byte, error) {
	return json.Marshal(req)
}

func jsonw(b []byte, res any) error {
	return json.Unmarshal(b, res)
}

func txtr(req any) ([]byte, error) {
	if b, ok := req.([]byte); ok {
		return b, nil
	}
	return nil, fmt.Errorf("invalid input type, need []byte")
}

func txtw(b []byte, res any) error {
	if rb, ok := res.([]byte); ok {
		rb = append(rb, b...)
		res = rb
		return nil
	}
	return fmt.Errorf("invalid putput type, need []byte")
}
