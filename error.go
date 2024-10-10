package springcloud

import "errors"

var (
	ErrorNoAvailableEndpoint = errors.New("no available endpoint")
)
