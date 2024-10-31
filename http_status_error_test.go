package springcloud

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpStatusErrorIs(t *testing.T) {
	err := NewHttpStatusError(http.StatusNotFound, "resource not found")
	var httpError = &HttpStatusError{
		StatusCode: http.StatusNotFound,
		StatusText: http.StatusText(http.StatusNotFound),
		Body:       "resource not found",
	}
	assert.True(t, errors.Is(err, httpError))
}

func TestHttpStatusErrorAs(t *testing.T) {
	err := NewHttpStatusError(http.StatusNotFound, "resource not found")
	var httpError *HttpStatusError
	assert.True(t, errors.As(err, &httpError))
}
