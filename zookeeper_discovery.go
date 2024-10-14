package springcloud

import (
	"encoding/json"
	"time"

	"github.com/libgox/gocollections/syncx"
	"github.com/protocol-laboratory/zookeeper-client-go/zk"
)

type ZooKeeperDiscoveryConfig struct {
	ZkConfig *zk.Config
}

type ZookeeperDiscovery struct {
	config *zk.Config

	client *zk.Client
	ticker *time.Ticker

	endpoints syncx.Map[string, []*Endpoint]
}

func NewZookeeperDiscovery(config *ZooKeeperDiscoveryConfig) (*ZookeeperDiscovery, error) {
	z := &ZookeeperDiscovery{
		config: config.ZkConfig,
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
			endpointsFromZk, err := z.getEndpointsFromZk(service)
			if err != nil {
				continue
			}
			z.endpoints.Store(service, endpointsFromZk)
		}
	}
}

func (z *ZookeeperDiscovery) getEndpointsFromZk(serviceName string) ([]*Endpoint, error) {
	path := "/services/" + serviceName
	resp, err := z.client.GetChildren(path)
	if err != nil {
		return nil, err
	}
	var endpointList []*Endpoint
	if len(resp.Children) == 0 {
		return endpointList, nil
	}
	for _, child := range resp.Children {
		var data *zk.GetDataResp
		data, err = z.client.GetData(path + "/" + child)
		if err != nil {
			return nil, err
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

func (z *ZookeeperDiscovery) Close() {
	z.ticker.Stop()
	z.client.Close()
}
