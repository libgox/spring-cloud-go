# spring-cloud-go

![License](https://img.shields.io/badge/license-Apache2.0-green)
![Language](https://img.shields.io/badge/Language-Go-blue.svg)
[![version](https://img.shields.io/github/v/tag/libgox/spring-cloud-go?label=release&color=blue)](https://github.com/libgox/spring-cloud-go/releases)
[![Godoc](http://img.shields.io/badge/docs-go.dev-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/libgox/spring-cloud-go)
[![Go report](https://goreportcard.com/badge/github.com/libgox/spring-cloud-go)](https://goreportcard.com/report/github.com/libgox/spring-cloud-go)
[![codecov](https://codecov.io/gh/libgox/spring-cloud-go/branch/main/graph/badge.svg)](https://codecov.io/gh/libgox/spring-cloud-go)

## 📋 Requirements

- Go 1.21+

## 🚀 Install

```
go get github.com/libgox/spring-cloud-go
```

## 💡 Usage

### Zookeeper Discovery Setup

```go
package main

import (
	"fmt"
	"log"

	"github.com/libgox/addr"
	springcloud "github.com/libgox/spring-cloud-go"
	"github.com/protocol-laboratory/zookeeper-client-go/zk"
)

func main() {
	config := &springcloud.ZooKeeperDiscoveryConfig{
		ZkConfig: &zk.Config{
			Addresses: []addr.Address{
				{
					Host: "localhost",
					Port: 2181,
				},
			},
		},
	}

	discovery, err := springcloud.NewZookeeperDiscovery(config)
	if err != nil {
		log.Fatalf("Failed to initialize discovery: %v", err)
	}

	endpoints, err := discovery.GetEndpoints("springcloud-service")
	if err != nil {
		log.Fatalf("Failed to get endpoints: %v", err)
	}

	fmt.Println("Discovered endpoints:", endpoints)
	defer discovery.Close()
}
```

### Spring Cloud Client Setup

The `spring_cloud_client` package in `spring-cloud-go` enables HTTP-based communication between services, leveraging Spring Cloud's service discovery mechanism.

#### Example Usage

```go
package main

import (
	"context"
	"fmt"
	"github.com/libgox/addr"
	"io/ioutil"
	"log"

	"crypto/tls"
	springcloud "github.com/libgox/spring-cloud-go"
	"github.com/protocol-laboratory/zookeeper-client-go/zk"
)

func main() {
	// Initialize ZooKeeper discovery
	discoveryConfig := &springcloud.ZooKeeperDiscoveryConfig{
		ZkConfig: &zk.Config{
			Addresses: []addr.Address{
				{
					Host: "localhost",
					Port: 2181,
				},
			},
		},
	}

	discovery, err := springcloud.NewZookeeperDiscovery(discoveryConfig)
	if err != nil {
		log.Fatalf("Failed to initialize discovery: %v", err)
	}
	defer discovery.Close()

	// Set up the client configuration
	clientConfig := springcloud.ClientConfig{
		Discovery: discovery,
		TlsConfig: &tls.Config{InsecureSkipVerify: true}, // Optionally configure TLS
	}

	// Create a new client
	client := springcloud.NewClient(clientConfig)

	// Make a GET request to a discovered service
	resp, err := client.Get(context.Background(), "springcloud-service", "/path/to/resource", nil)
	if err != nil {
		log.Fatalf("Failed to perform GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Println("Response:", string(body))
}
```

#### Key Features:
- **Service Discovery Integration**: Automatically discover services registered in Zookeeper.
- **Load Balancing**: Client requests are distributed across available endpoints using a round-robin strategy.
- **Flexible HTTP Methods**: Supports `GET`, `POST`, `PUT`, and `DELETE` HTTP methods for interacting with services.
- **TLS Support**: Optional TLS configuration for secure service communication.

### Available Methods
- `client.Get`: Sends an HTTP `GET` request.
- `client.Post`: Sends an HTTP `POST` request with a body.
- `client.Put`: Sends an HTTP `PUT` request with a body.
- `client.Delete`: Sends an HTTP `DELETE` request.
