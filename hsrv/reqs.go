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

type RequestHandler func(Response, Request)

type ErrorHandler func(Response, Request, error)

// Response=========================================================

type Response struct {
	http.ResponseWriter
	Ctx *reqCtx
}

func (r *Response) doResponse(code int, body []byte) (err error) {
	if r.Ctx.after != nil {
		for _, pp := range r.Ctx.after {
			pp.After(*r)
		}
	}

	r.WriteHeader(code)
	_, err = r.Write(body)
	return
}

func (r *Response) doResponse1(code int, body io.Reader) (wirtten int64, err error) {

	r.WriteHeader(http.StatusOK)
	return io.Copy(r, body)

}

func (r *Response) Text(txt string, statusCode int) (err error) {
	r.Header().Add("Content-Type", "text/plain; charset=utf-8")
	r.doResponse(statusCode, util.String2Bytes(txt))
	return
}

func (r *Response) Json(target any, statusCode int) error {
	b, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("error on unmarshal json, %w", err)
	}
	r.Header().Add("Content-Type", "application/json; charset=utf-8")
	r.doResponse(statusCode, b)
	return err
}

func (r *Response) File(input io.Reader, size int64, name string) (int64, error) {
	ct, down := getContentTypeByFilename(name)
	cd := ""
	if down {
		cd = "attachment; filename=" + name
	} else {
		cd = " filename=" + name
	}

	r.Header().Set("Content-Disposition", cd)
	r.Header().Set("Content-Type", ct)
	if size > 0 {
		r.Header().Set("Content-Length", strconv.FormatInt(size, 10))
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
	}

	return "application/octet-stream", true

}

// Request=============================================================

type Request struct {
	*http.Request
	Ctx *reqCtx
}

func (r *Request) IP() (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}

func (r *Request) WebContext() any {
	return r.Context().Value(WebContextVName)
}

func (r *Request) Text() (string, error) {
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("error on read request body, %w", err)
	}
	return string(b), nil
}

func (r *Request) Json(target any) error {
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error on read request body, %w", err)
	}

	err = json.Unmarshal(b, target)
	if err != nil {
		return fmt.Errorf("error on unmarshal json, %w", err)
	}
	return nil
}

func (r *Request) UrlParam() (map[string]string, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, r.ContentLength))
	buffer.ReadFrom(r.Body)
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

func (r *Request) Stream(wr io.Writer) (err error) {
	defer r.Body.Close()
	_, err = io.Copy(wr, r.Body)
	return
}

func (r *Request) FormData() (map[string][]string, error) {
	defer r.Body.Close()
	contentType := r.Header.Get("content-type")

	var formValue map[string][]string = make(map[string][]string)
	if !strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return formValue, fmt.Errorf("content-type must be application/x-www-form-urlencoded")
	}

	err := r.ParseForm()
	if err != nil {
		return formValue, fmt.Errorf("error on pares form data, %w", err)
	}

	return r.Form, nil
}

type MultiForm struct {
	close bool
	*multipart.Form
	r *Request
}

func (m *MultiForm) Close() {
	if m.close {
		m.r.Body.Close()
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

func (r *Request) MultipartForm() (f *MultiForm, err error) {
	return r.MultipartFormLarge(16)
}

// 处理更大的formdata
// size: 单位为 Mib
func (r *Request) MultipartFormLarge(size int64) (f *MultiForm, err error) {
	contentType := r.Header.Get("content-type")

	if !strings.Contains(contentType, "multipart/form-data") {
		err = fmt.Errorf("content-type must be multipart/form-data")
		return
	}

	var err2 error = r.ParseMultipartForm(2 * size * 1024 * 1024)
	if err2 != nil {
		err = fmt.Errorf("failure on parse files, %w", err2)
		return
	}
	f = &MultiForm{
		close: false,
		Form:  r.Request.MultipartForm,
		r:     r,
	}

	return
}
