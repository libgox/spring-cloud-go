package springcloud

import (
	"encoding/json"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/libgox/gocollections/syncx"
	"github.com/protocol-laboratory/zookeeper-client-go/zk"
)

type ZooKeeperDiscoveryConfig struct {
	ZkConfig *zk.Config

	// Logger structured logger for logging operations
	Logger *slog.Logger
}

var _ Discovery = (*ZookeeperDiscovery)(nil)

type ZookeeperDiscovery struct {
	config *zk.Config

	client *zk.Client
	ticker *time.Ticker

	endpoints syncx.Map[string, []*Endpoint]

	logger *slog.Logger
}

func NewZookeeperDiscovery(config *ZooKeeperDiscoveryConfig) (*ZookeeperDiscovery, error) {
	z := &ZookeeperDiscovery{
		config: config.ZkConfig,
	}
	if config.Logger != nil {
		z.logger = config.Logger
	} else {
		z.logger = slog.Default()
	}
	var err error
	z.client, err = zk.NewClient(z.config)
	if err != nil {
		return nil, err
	}
	z.ticker = time.NewTicker(30 * time.Second)
	go func() {
		z.updateEndpoints()
	}()
	return z, nil
}

func (z *ZookeeperDiscovery) GetEndpoints(serviceName string) ([]*Endpoint, error) {
	value, ok := z.endpoints.Load(serviceName)
	if ok {
		return value, nil
	}
	endpoints, err := z.getEndpointsFromZk(serviceName)
	if err != nil {
		return nil, err
	}
	z.endpoints.Store(serviceName, endpoints)
	return endpoints, nil
}

func (z *ZookeeperDiscovery) updateEndpoints() {
	for range z.ticker.C {
		services := make([]string, 0)
		z.endpoints.Range(func(key string, value []*Endpoint) bool {
			services = append(services, key)
			return true
		})
		for _, service := range services {
			z.logger.Info("fetching new endpoints from zookeeper", slog.String("service", service))
			endpointsFromZk, err := z.getEndpointsFromZk(service)
			if err != nil {
				z.logger.Error("failed to fetch endpoints from zookeeper", slog.String("service", service), slog.Any("error", err))
				continue
			}
			z.logger.Info("successfully fetched endpoints", slog.String("service", service), slog.String("ips", formatIPs(extractEndpointIPs(endpointsFromZk))))
			z.endpoints.Store(service, endpointsFromZk)
		}
	}
}

func extractEndpointIPs(endpoints []*Endpoint) []string {
	var ips []string
	for _, endpoint := range endpoints {
		ips = append(ips, endpoint.Address)
	}
	return ips
}

func formatIPs(ips []string) string {
	return strings.Join(ips, ", ")
}

func (z *ZookeeperDiscovery) getEndpointsFromZk(serviceName string) ([]*Endpoint, error) {
	path := "/services/" + serviceName
	z.logger.Debug("fetching endpoints from zookeeper", slog.String("path", path))
	resp, err := z.client.GetChildren(path)
	if err != nil {
		return nil, err
	}
	z.logger.Debug("fetched endpoints from zookeeper", slog.String("path", path), slog.Any("endpoints", resp.Children))
	var endpointList []*Endpoint
	if len(resp.Children) == 0 {
		return endpointList, nil
	}
	for _, child := range resp.Children {
		z.logger.Debug("fetching endpoint from zookeeper", slog.String("path", path+"/"+child))
		var data *zk.GetDataResp
		data, err = z.client.GetData(path + "/" + child)
		if err != nil {
			return nil, err
		}
		z.logger.Debug("fetched endpoint from zookeeper", slog.String("path", path+"/"+child), slog.Any("data", data))
		var endpoint Endpoint
		err = json.Unmarshal(data.Data, &endpoint)
		if err != nil {
			return nil, err
		}
		endpointList = append(endpointList, &endpoint)
	}
	return endpointList, nil
}

func (z *ZookeeperDiscovery) Close() error {
	z.ticker.Stop()
	z.client.Close()
	return nil
}
