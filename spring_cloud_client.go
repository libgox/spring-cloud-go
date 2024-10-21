package springcloud

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/textproto"
	"sync/atomic"

	"github.com/libgox/gocollections/syncx"
)

type ClientConfig struct {
	Discovery Discovery
	TlsConfig *tls.Config
}

type Client struct {
	discovery  Discovery
	httpClient *http.Client
	tlsEnabled bool
	rrIndices  syncx.Map[string, *atomic.Uint32]
}

func NewClient(config ClientConfig) *Client {
	transport := &http.Transport{
		TLSClientConfig: config.TlsConfig,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	return &Client{
		discovery:  config.Discovery,
		httpClient: httpClient,
		tlsEnabled: config.TlsConfig != nil,
	}
}

func (c *Client) Get(ctx context.Context, serviceName string, path string, headers textproto.MIMEHeader) (*http.Response, error) {
	return c.Request(ctx, serviceName, http.MethodGet, path, nil, headers)
}

func (c *Client) Post(ctx context.Context, serviceName string, path string, body []byte, headers textproto.MIMEHeader) (*http.Response, error) {
	return c.Request(ctx, serviceName, http.MethodPost, path, body, headers)
}

func (c *Client) Put(ctx context.Context, serviceName string, path string, body []byte, headers textproto.MIMEHeader) (*http.Response, error) {
	return c.Request(ctx, serviceName, http.MethodPut, path, body, headers)
}

func (c *Client) Delete(ctx context.Context, serviceName string, path string, headers textproto.MIMEHeader) (*http.Response, error) {
	return c.Request(ctx, serviceName, http.MethodDelete, path, nil, headers)
}

func (c *Client) Request(ctx context.Context, serviceName string, method string, path string, body []byte, headers textproto.MIMEHeader) (*http.Response, error) {
	endpoints, err := c.discovery.GetEndpoints(serviceName)
	if err != nil {
		return nil, err
	}

	endpoint, ok := c.getNextEndpoint(serviceName, endpoints)
	if !ok {
		return nil, ErrorNoAvailableEndpoint
	}

	var prefix string
	if c.tlsEnabled {
		prefix = "https://"
	} else {
		prefix = "http://"
	}
	url := fmt.Sprintf("%s%s:%d%s", prefix, endpoint.Address, endpoint.Port, path)

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %v", err)
	}

	return resp, nil
}

func (c *Client) getNextEndpoint(serviceName string, endpoints []*Endpoint) (*Endpoint, bool) {
	if len(endpoints) == 0 {
		return nil, false
	}

	_, ok := c.rrIndices.Load(serviceName)
	if !ok {
		var newRRIndex atomic.Uint32
		c.rrIndices.Store(serviceName, &newRRIndex)
	}

	// load rrIndex again
	rrIndex, ok := c.rrIndices.Load(serviceName)
	if !ok {
		return nil, false
	}

	index := rrIndex.Load()
	nextIndex := (index + 1) % uint32(len(endpoints))
	rrIndex.Store(nextIndex)

	return endpoints[index], true
}
