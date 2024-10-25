package springcloud

import "errors"

var (
	ErrNoAvailableEndpoint = errors.New("no available endpoint")
)
