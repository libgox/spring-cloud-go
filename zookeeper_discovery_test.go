package springcloud

import (
	"github.com/protocol-laboratory/zookeeper-client-go/zk"
	"github.com/shoothzj/gox/netx"
	"github.com/stretchr/testify/assert"
	"testing"
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
