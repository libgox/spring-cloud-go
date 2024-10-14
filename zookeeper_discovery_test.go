package springcloud

import (
	"testing"

	"github.com/protocol-laboratory/zookeeper-client-go/zk"
	"github.com/shoothzj/gox/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZookeeperDiscovery(t *testing.T) {
	zd, err := NewZookeeperDiscovery(&ZooKeeperDiscoveryConfig{
		ZkConfig: &zk.Config{
			Addresses: []netx.Address{
				{
					Host: "localhost",
					Port: 2181,
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, zd)
	assert.NotNil(t, zd.client)
	assert.NotNil(t, zd.ticker)
	zd.Close()
}

func TestGetEndpointZooKeeperData(t *testing.T) {
	config := &zk.Config{
		Addresses: []netx.Address{
			{
				Host: "localhost",
				Port: 2181,
			},
		},
	}
	client, err := zk.NewClient(config)
	require.NoError(t, err)
	defer client.Close()
	_, err = client.Create("/services", []byte{}, []int{31}, "world", "anyone", 0)
	require.NoError(t, err)
	//nolint:errcheck
	defer client.Delete("/services", -1)
	_, err = client.Create("/services/test", []byte{}, []int{31}, "world", "anyone", 0)
	require.NoError(t, err)
	//nolint:errcheck
	defer client.Delete("/services/test", -1)
	_, err = client.Create("/services/test/id1", []byte(`
{
    "name":"service",
    "id":"id",
    "address":"localhost",
    "port":8080,
    "sslPort":null,
    "registrationTimeUTC":0,
    "serviceType":"DYNAMIC"
}
`), []int{31}, "world", "anyone", 0)
	require.NoError(t, err)
	//nolint:errcheck
	defer client.Delete("/services/test/id1", -1)
	zd, err := NewZookeeperDiscovery(&ZooKeeperDiscoveryConfig{
		ZkConfig: &zk.Config{
			Addresses: []netx.Address{
				{
					Host: "localhost",
					Port: 2181,
				},
			},
		},
	})
	require.NoError(t, err)
	defer zd.Close()
	endpoints, err := zd.GetEndpoints("test")
	require.NoError(t, err)
	assert.Len(t, endpoints, 1)
	endpoint := endpoints[0]
	assert.Equal(t, "service", endpoint.Name)
	assert.Equal(t, "id", endpoint.Id)
	assert.Equal(t, "localhost", endpoint.Address)
	assert.Equal(t, 8080, endpoint.Port)
	assert.Equal(t, (*int)(nil), endpoint.SslPort)
	assert.Equal(t, int64(0), endpoint.RegistrationTimeUTC)
	assert.Equal(t, "DYNAMIC", endpoint.ServiceType)
}
