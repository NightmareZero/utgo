package hsrv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/NightmareZero/nzgoutil/util"
)

type RequestHandler func(Ctx)

type ErrorHandler func(Ctx, error)

type Ctx struct {
	w http.ResponseWriter
	r *http.Request
	*reqCtx
}

// Response=========================================================

func (r *Ctx) doResponse(code int, body []byte) (err error) {
	if r.after != nil {
		for _, pp := range r.after {
			pp.After(*r)
		}
	}

	r.w.WriteHeader(code)
	_, err = r.w.Write(body)
	return
}

func (r *Ctx) doResponse1(code int, body io.Reader) (wirtten int64, err error) {

	r.w.WriteHeader(http.StatusOK)
	return io.Copy(r.w, body)

}

func (r *Ctx) Text(statusCode int, txt string) (err error) {
	r.w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	r.doResponse(statusCode, util.String2Bytes(txt))
	return
}

func (r *Ctx) Html(statusCode int, html string) (err error) {
	r.w.Header().Add("Content-Type", "text/html; charset=utf-8")
	r.doResponse(statusCode, util.String2Bytes(html))
	return
}

func (r *Ctx) Sse(fun func(hf http.Flusher, r *Ctx)) (err error) {
	flusher, ok := r.w.(http.Flusher)
	if !ok {
		http.Error(r.w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	r.w.Header().Add("Content-Type", "text/event-stream; charset=utf-8")
	r.w.Header().Set("Cache-Control", "no-cache")
	r.w.Header().Set("Connection", "keep-alive")
	r.w.Header().Set("Access-Control-Allow-Origin", "*")

	// fun 方法说明:
	// 返回数据包含id、event(非必须,如果包含本项，则必须使用eventlistener接收，否则为onMessage)、data，结尾必须使用\n\n
	// fmt.Fprintf(rw, "id: %d\nevent: ping \ndata: %d\n\n", time.Now().Unix(), time.Now().Unix())
	// flusher.Flush()
	fun(flusher, r)
	return
}

func (r *Ctx) Json(statusCode int, target any) error {
	b, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("error on unmarshal json, %w", err)
	}
	r.w.Header().Add("Content-Type", "application/json; charset=utf-8")
	r.doResponse(statusCode, b)
	return err
}

func (r *Ctx) File(input io.Reader, size int64, name string) (int64, error) {
	ct, down := getContentTypeByFilename(name)
	cd := ""
	if down {
		cd = "attachment; filename=" + name
	} else {
		cd = " filename=" + name
	}

	r.w.Header().Set("Content-Disposition", cd)
	r.w.Header().Set("Content-Type", ct)
	if size > 0 {
		r.w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}

	w, err := r.doResponse1(http.StatusOK, input)
	if err != nil {
		return w, fmt.Errorf("fail on response file, %w", err)
	}

	return w, nil
}

func getContentTypeByFilename(name string) (ct string, down bool) {
	lName := strings.ToLower(name)
	s := path.Ext(lName)

	switch s {
	case ".jpg":
		fallthrough
	case ".jpeg":
		return "image/jpeg", false
	case ".png":
		return "image/png", false
	case ".svg":
		return "image/svg+xml", false
	case ".ico":
		return "image/x-icon", false
	case ".tif":
		return "image/tiff", false
	case ".gif":
		return "image/gif", false
	case ".bmp":
		return "image/bmp", false
	case ".webp":
		return "image/webp", false
	case ".mp4":
		return "video/mp4", false
	case ".avi":
		return "video/x-msvideo", false
	case ".mov":
		return "video/quicktime", false
	case ".mp3":
		return "audio/mpeg", false
	case ".wav":
		return "audio/wav", false
	case ".ogg":
		return "audio/ogg", false
	case ".pdf":
		return "application/pdf", false
	}

	return "application/octet-stream", true

}

// Request=============================================================

func (r *Ctx) IP() (string, error) {
	ip := r.r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	ip = r.r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}

func (r *Ctx) WebContext() any {
	return r.r.Context().Value(WebContextVName)
}

func (r *Ctx) ParseText() (string, error) {
	defer r.r.Body.Close()
	b, err := io.ReadAll(r.r.Body)
	if err != nil {
		return "", fmt.Errorf("error on read request body, %w", err)
	}
	return string(b), nil
}

func (r *Ctx) ParseJson(target any) error {
	defer r.r.Body.Close()

	b, err := io.ReadAll(r.r.Body)
	if err != nil {
		return fmt.Errorf("error on read request body, %w", err)
	}

	err = json.Unmarshal(b, target)
	if err != nil {
		return fmt.Errorf("error on unmarshal json, %w", err)
	}
	return nil
}

func (r *Ctx) UrlParam() (map[string]string, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, r.r.ContentLength))
	buffer.ReadFrom(r.r.Body)
	fmt.Println(buffer.String())

	params, err := url.ParseQuery(buffer.String())
	if err != nil {
		return nil, fmt.Errorf("error on parse url param, %w", err)
	}
	values := map[string]string{}
	for key, value := range params {
		if len(value) > 0 {
			values[key] = value[0]
		} else {
			values[key] = ""
		}
	}
	return values, nil
}

func (r *Ctx) Stream(wr io.Writer) (err error) {
	defer r.r.Body.Close()
	_, err = io.Copy(wr, r.r.Body)
	return
}

func (r *Ctx) FormData() (map[string][]string, error) {
	defer r.r.Body.Close()
	contentType := r.r.Header.Get("content-type")

	var formValue map[string][]string = make(map[string][]string)
	if !strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return formValue, fmt.Errorf("content-type must be application/x-www-form-urlencoded")
	}

	err := r.r.ParseForm()
	if err != nil {
		return formValue, fmt.Errorf("error on pares form data, %w", err)
	}

	return r.r.Form, nil
}

type MultiForm struct {
	close bool
	*multipart.Form
	r *Ctx
}

func (m *MultiForm) Close() {
	if m.close {
		m.r.r.Body.Close()
		m.close = true
	}
}

func (m *MultiForm) ParseFile(writer func(string, io.Reader, error)) {
	for _, content := range m.File {
		for _, fh := range content {
			f, err2 := fh.Open()
			if err2 != nil {
				writer("", nil, err2)
			}
			writer(fh.Filename, f, nil)
			f.Close()
		}
	}
	m.Close()
}

func (r *Ctx) MultipartForm() (f *MultiForm, err error) {
	return r.MultipartFormLarge(16)
}

// 处理更大的formdata
// size: 单位为 Mib
func (r *Ctx) MultipartFormLarge(size int64) (f *MultiForm, err error) {
	contentType := r.r.Header.Get("content-type")

	if !strings.Contains(contentType, "multipart/form-data") {
		err = fmt.Errorf("content-type must be multipart/form-data")
		return
	}

	var err2 error = r.r.ParseMultipartForm(2 * size * 1024 * 1024)
	if err2 != nil {
		err = fmt.Errorf("failure on parse files, %w", err2)
		return
	}
	f = &MultiForm{
		close: false,
		Form:  r.r.MultipartForm,
		r:     r,
	}

	return
}
