package seaweedfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// errSmallFileAppend 标识 Filer 对内联小文件追加的限制。
// 源码: filer_server_handlers_write_autochunk.go 的 saveMetaData 中
//
//	if len(entry.Content) > 0 { return "append to small file is not supported yet" }
var errSmallFileAppendPrefix = "append to small file is not supported"

// filerClient 封装 SeaweedFS Filer HTTP 调用，用于追加写。
// 当 baseURL 为空时，追加功能降级为 read-modify-write（由调用方处理）。
type filerClient struct {
	baseURL string          // 含 scheme，如 "http://localhost:8888"；为空表示不可用
	hc      *http.Client    // 复用的 HTTP 客户端
	access  string          // 可选凭证（用于 V4 签名，当前默认无认证）
	secret  string
}

func newFilerClient(baseURL, access, secret string) *filerClient {
	return &filerClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		hc:      &http.Client{Timeout: 5 * time.Minute},
		access:  access,
		secret:  secret,
	}
}

// available 返回 FilerURL 是否已配置（追加可走原生 ?op=append）。
func (c *filerClient) available() bool {
	return c != nil && c.baseURL != ""
}

// append 将 p 追加到指定对象末尾，走 Filer ?op=append。
// 返回 isSmallFileErr=true 表示命中"小文件不支持追加"限制，调用方应回退 read-modify-write。
func (c *filerClient) append(ctx context.Context, bucket, name string, p []byte) (isSmallFileErr bool, err error) {
	if !c.available() {
		return false, errors.New("filerClient: baseURL not configured")
	}

	// 构造 multipart/form-data body（Filer POST 期望文件以 multipart part 上传）
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile("file", name)
	if err != nil {
		return false, err
	}
	if _, err := fw.Write(p); err != nil {
		return false, err
	}
	if err := mw.Close(); err != nil {
		return false, err
	}

	url := fmt.Sprintf("%s/%s/%s?op=append", c.baseURL, bucket, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	// 注：SeaweedFS Filer HTTP 默认无认证；如部署启用，此处可补充 AWS V4 签名头

	resp, err := c.hc.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 读取响应体用于错误判断
	respBody, _ := io.ReadAll(resp.Body)
	respText := string(respBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return false, nil
	}

	// 命中小文件限制（Filer 返回 500 + 错误消息）
	if strings.Contains(respText, errSmallFileAppendPrefix) {
		return true, fmt.Errorf("filer append small file not supported: %s", respText)
	}
	return false, fmt.Errorf("filer append failed: status=%d body=%s", resp.StatusCode, respText)
}

// readAll 读取整个对象内容，供 read-modify-write 回退使用。
func (c *filerClient) readAll(ctx context.Context, bucket, name string) ([]byte, error) {
	if !c.available() {
		return nil, errors.New("filerClient: baseURL not configured")
	}
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, bucket, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("filer read failed: status=%d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
