package hsrv

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Request struct {
	*http.Request
}

type Response struct {
	http.ResponseWriter
}

type RequestHandler func(Response, Request)

type ErrorHandler func(Response, Request, error)

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

func (r *Request) Stream(wr io.Writer) (err error) {
	defer r.Body.Close()
	_, err = io.Copy(wr, r.Body)
	return
}

func (r *Response) Text(txt string, statusCode int) (err error) {
	r.WriteHeader(statusCode)
	r.Header().Add("Content-Type", "text/plain; charset=utf-8")
	_, err = r.Write([]byte(txt))
	return
}

func (r *Response) Json(target any, statusCode int) error {
	r.WriteHeader(statusCode)
	b, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("error on unmarshal json, %w", err)
	}
	r.Header().Add("Content-Type", "application/json; charset=utf-8")
	_, err = r.Write([]byte(b))
	return err
}

func (r *Request) FormData(writer func(string, io.Reader, error), multiThread bool) (map[string][]string, error) {
	contentType := r.Header.Get("content-type")
	contentLen := r.ContentLength

	var formValue map[string][]string = make(map[string][]string)
	if !strings.Contains(contentType, "multipart/form-data") {
		return formValue, fmt.Errorf("content-type must be multipart/form-data")
	}

	var err error
	if contentLen < 4*1024*1024 {
		err = r.ParseMultipartForm(4 * 1024 * 1024)
	} else {
		err = r.ParseMultipartForm(16 * 1024 * 1024)
	}
	if err != nil {
		return formValue, fmt.Errorf("failure on parse files")
	}

	if len(r.MultipartForm.File) == 0 {
		return formValue, fmt.Errorf("not have any file")
	}

	for k, v := range r.MultipartForm.Value {
		formValue[k] = v
	}

	for name, content := range r.MultipartForm.File {
		if multiThread {
			go uploadFilesWorker(name, content, writer)
		} else {
			uploadFilesWorker(name, content, writer)
		}

	}
	return formValue, nil
}

func uploadFilesWorker(name string, cc []*multipart.FileHeader,
	writer func(string, io.Reader, error)) {
	for _, fh := range cc {
		f, err2 := fh.Open()
		if err2 != nil {
			writer("", nil, err2)
		}
		writer(fh.Filename, f, nil)
		f.Close()
	}

}
