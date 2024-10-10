package springcloud

type Discovery interface {
	GetEndpoints(serviceName string) ([]*Endpoint, error)
}
