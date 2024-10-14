# spring-cloud-go

![License](https://img.shields.io/badge/license-Apache2.0-green) ![Language](https://img.shields.io/badge/Language-Go-blue.svg)

## Requirements

- Go 1.20+

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
	springcloud "github.com/libgox/spring-cloud-go"
	"github.com/protocol-laboratory/zookeeper-client-go/zk"
	"github.com/shoothzj/gox/netx"
	"log"
)

func main() {
	config := &springcloud.Config{
		ZkConfig: &zk.Config{
			Addresses: []netx.Address{
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
