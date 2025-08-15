package discovery

import (
	"context"

	"github.com/ripplego/ripplego/internal/core"
)

// Finder 抽象节点发现接口，支持 mDNS、DHT 等实现
type Finder interface {
	Start(ctx context.Context) error
	Stop() error
	Nodes() []core.Node
}