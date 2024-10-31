package springcloud

import (
	"fmt"
	"net/http"
)

type HttpStatusError struct {
	StatusCode int
	StatusText string
	Body       string
}

func (e *HttpStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s - %s", e.StatusCode, e.StatusText, e.Body)
}

func NewHttpStatusError(statusCode int, body string) *HttpStatusError {
	return &HttpStatusError{
		StatusCode: statusCode,
		StatusText: http.StatusText(statusCode),
		Body:       body,
	}
}
