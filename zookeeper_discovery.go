package springcloud

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

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
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	z := &ZookeeperDiscovery{
		config: config.ZkConfig,
		logger: config.Logger,
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
		func() {
			defer func() {
				if r := recover(); r != nil {
					var buf [4096]byte
					n := runtime.Stack(buf[:], false)
					stackInfo := string(buf[:n])
					z.logger.Error(fmt.Sprintf("%v cause panic, stack: %s", r, stackInfo))
				}
			}()

			services := make([]string, 0)
			z.endpoints.Range(func(key string, value []*Endpoint) bool {
				services = append(services, key)
				return true
			})

			for _, service := range services {
				z.logger.Info("fetching new endpoints from zookeeper", slog.String(LogKeyService, service))
				endpointsFromZk, err := z.getEndpointsFromZk(service)
				if err != nil {
					z.logger.Error("failed to fetch endpoints from zookeeper", slog.String(LogKeyService, service), slog.Any(LogKeyError, err))
					continue
				}
				z.logger.Info("successfully fetched endpoints", slog.String(LogKeyService, service), slog.String(LogKeyIps, formatIPs(extractEndpointIPs(endpointsFromZk))))
				z.endpoints.Store(service, endpointsFromZk)
			}
		}()
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
	z.logger.Debug("fetching endpoints from zookeeper", slog.String(LogKeyPath, path))
	resp, err := z.client.GetChildren(path)
	if err != nil {
		return nil, err
	}
	z.logger.Debug("fetched endpoints from zookeeper", slog.String(LogKeyPath, path), slog.Any(LogKeyEndpoints, resp.Children))
	var endpointList []*Endpoint
	if len(resp.Children) == 0 {
		return endpointList, nil
	}
	for _, child := range resp.Children {
		z.logger.Debug("fetching from zookeeper", slog.String(LogKeyPath, path+"/"+child))
		var data *zk.GetDataResp
		data, err = z.client.GetData(path + "/" + child)
		if err != nil {
			return nil, err
		}
		z.logger.Debug("fetched data", slog.String(LogKeyPath, path+"/"+child), slog.Any(LogKeyData, data))
		if data.Error == zk.EcNoNode {
			z.logger.Info("node not found", slog.String(LogKeyPath, path+"/"+child))
			continue
		}
		if data.Error != zk.EcOk {
			return nil, fmt.Errorf("failed to get data from zookeeper: %v", data.Error)
		}
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
