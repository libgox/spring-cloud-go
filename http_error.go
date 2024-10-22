package springcloud

import (
	"errors"
	"fmt"
	"net/http"
)

type HttpStatusError struct {
	StatusCode int
	StatusText string
	Body       string
}

func (e *HttpStatusError) Is(target error) bool {
	var t *HttpStatusError
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.StatusCode == t.StatusCode && e.Body == t.Body
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
