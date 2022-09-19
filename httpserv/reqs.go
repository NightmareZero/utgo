package httpserv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Request http.Request

type Response struct {
	writer http.ResponseWriter
}

type RequestHandler func(Request, Response)

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

func (r *Response) Status(s int) {
	r.writer.WriteHeader(s)
}

func (r *Response) Headers() http.Header {
	return r.writer.Header()
}

func (r *Response) Text(txt string) (err error) {
	r.writer.WriteHeader(http.StatusOK)
	r.writer.Header().Add("Content-Type", "text/plain")
	_, err = r.writer.Write([]byte(txt))
	return
}

func (r *Response) Json(target any) error {
	b, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("error on unmarshal json, %w", err)
	}
	r.writer.WriteHeader(http.StatusOK)
	r.writer.Header().Add("Content-Type", "application/json")
	_, err = r.writer.Write([]byte(b))
	return err
}
