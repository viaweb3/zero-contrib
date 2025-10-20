# Apollo Config Center Integration

Apollo configuration center integration for go-zero framework.

## Installation

```bash
go get github.com/zeromicro/zero-contrib/configcenter/apollo
```

## Quick Start

```go
package main

import (
    "fmt"

    configurator "github.com/zeromicro/go-zero/core/configcenter"
    "github.com/zeromicro/zero-contrib/configcenter/apollo"
)

type AppConfig struct {
    Name    string `json:"name"`
    Timeout int64  `json:"timeout"`
}

func main() {
    // Create Apollo subscriber
    sub := apollo.MustNewApolloSubscriber(apollo.ApolloConf{
        AppID:         "your-app-id",
        Cluster:       "default",
        NamespaceName: "application.json",
        MetaAddr:      "http://localhost:8080",
        Format:        "json",
    })

    // Create config center
    cc := configurator.MustNewConfigCenter[AppConfig](
        configurator.Config{Type: "json"},
        sub,
    )

    // Get config
    config, _ := cc.GetConfig()
    fmt.Printf("Config: %+v\n", config)

    // Listen for changes
    cc.AddListener(func() {
        newConfig, _ := cc.GetConfig()
        fmt.Printf("Config updated: %+v\n", newConfig)
    })

    select {} // Keep running
}
```

## Configuration

### ApolloConf

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| AppID | string | Yes | - | Apollo application ID |
| MetaAddr | string | Yes | - | Apollo meta server address |
| Cluster | string | No | "default" | Cluster name |
| NamespaceName | string | No | "application" | Namespace name |
| Format | string | No | "json" | Config format: json/yaml/properties |
| Key | string | No | - | Specific key to watch (empty = entire namespace) |
| Secret | string | No | - | Secret key for authentication |
| IsBackupConfig | bool | No | false | Enable local backup |
| BackupPath | string | No | - | Backup directory path |

## Examples

See [examples](./examples) directory for more examples.

## References

- [Apollo Documentation](https://www.apolloconfig.com/)
- [go-zero Config Center](https://go-zero.dev/docs/tutorials/configcenter/overview)
