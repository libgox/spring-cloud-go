package springcloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_getNextEndpoint_NoEndpoints(t *testing.T) {
	client := &Client{}

	endpoint, ok := client.getNextEndpoint("testService", []*Endpoint{})

	assert.False(t, ok, "Expected ok to be false when endpoints list is empty")
	assert.Nil(t, endpoint, "Expected nil endpoint when endpoints list is empty")
}

func TestClient_getNextEndpoint_SingleEndpoint(t *testing.T) {
	client := &Client{}

	endpoint1 := &Endpoint{Address: "endpoint1"}
	endpoints := []*Endpoint{endpoint1}

	endpoint, ok := client.getNextEndpoint("testService", endpoints)

	assert.True(t, ok, "Expected ok to be true with a single endpoint")
	assert.Equal(t, endpoint1, endpoint, "Expected the single endpoint to be returned")
}

func TestClient_getNextEndpoint_MultipleEndpointsRoundRobin(t *testing.T) {
	client := &Client{}

	endpoint1 := &Endpoint{Address: "endpoint1"}
	endpoint2 := &Endpoint{Address: "endpoint2"}
	endpoint3 := &Endpoint{Address: "endpoint3"}
	endpoints := []*Endpoint{endpoint1, endpoint2, endpoint3}

	endpoint, ok := client.getNextEndpoint("testService", endpoints)
	assert.True(t, ok)
	assert.Equal(t, endpoint1, endpoint, "Expected first call to return first endpoint")

	endpoint, ok = client.getNextEndpoint("testService", endpoints)
	assert.True(t, ok)
	assert.Equal(t, endpoint2, endpoint, "Expected second call to return second endpoint")

	endpoint, ok = client.getNextEndpoint("testService", endpoints)
	assert.True(t, ok)
	assert.Equal(t, endpoint3, endpoint, "Expected third call to return third endpoint")

	endpoint, ok = client.getNextEndpoint("testService", endpoints)
	assert.True(t, ok)
	assert.Equal(t, endpoint1, endpoint, "Expected fourth call to loop back to first endpoint")
}
