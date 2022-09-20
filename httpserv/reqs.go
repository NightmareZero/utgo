package httpserv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	*http.Request
}

type Response struct {
	http.ResponseWriter
}

type RequestHandler func(Response, Request)

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

func (r *Response) Status(s int) {
	r.WriteHeader(s)
}

func (r *Response) Headers() http.Header {
	return r.Header()
}

func (r *Response) Text(txt string) (err error) {
	r.WriteHeader(http.StatusOK)
	r.Header().Add("Content-Type", "text/plain")
	_, err = r.Write([]byte(txt))
	return
}

func (r *Response) Json(target any) error {
	b, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("error on unmarshal json, %w", err)
	}
	r.WriteHeader(http.StatusOK)
	r.Header().Add("Content-Type", "application/json")
	_, err = r.Write([]byte(b))
	return err
}
