package discovery

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	mdns "github.com/hashicorp/mdns"
	"github.com/ripplego/ripplego/internal/core"
)

const (
	serviceName = "_ripplego._tcp"
)

// MDNSFinder 基于 mDNS 的局域网节点发现
// 通过广播 serviceName 公布本节点，监听并收集其它节点的信息

type MDNSFinder struct {
	port   int
	host   string
	mu     sync.RWMutex
	nodes  map[core.NodeID]core.Node
	ctx    context.Context
	cancel context.CancelFunc
	server *mdns.Server
}

func NewMDNSFinder(host string, port int) *MDNSFinder {
	return &MDNSFinder{
		port:  port,
		host:  host,
		nodes: make(map[core.NodeID]core.Node),
	}
}

func (m *MDNSFinder) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	info := []string{"RippleGo node"}
	service, err := mdns.NewMDNSService(m.host, serviceName, "local.", "", m.port, nil, info)
	if err != nil {
		return err
	}
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return err
	}
	m.server = server

	go m.queryLoop()
	return nil
}

func (m *MDNSFinder) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}
	if m.server != nil {
		m.server.Shutdown()
	}
	return nil
}

func (m *MDNSFinder) Nodes() []core.Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]core.Node, 0, len(m.nodes))
	for _, n := range m.nodes {
		out = append(out, n)
	}
	return out
}

func (m *MDNSFinder) queryLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.lookup()
		}
	}
}

func (m *MDNSFinder) lookup() {
	entries := make(chan *mdns.ServiceEntry, 16)
	go func() {
		for e := range entries {
			addr := fmt.Sprintf("%s:%d", e.AddrV4, e.Port)
			if e.AddrV4 == nil {
				if ip := net.IP(e.Addr); ip != nil {
					addr = fmt.Sprintf("%s:%d", ip.String(), e.Port)
				} else {
					continue
				}
			}
			node := core.Node{
				ID:       core.NodeID(e.Host),
				Address:  addr,
				LastSeen: time.Now(),
				Status:   "online",
			}
			m.mu.Lock()
			m.nodes[node.ID] = node
			m.mu.Unlock()
		}
	}()
	_ = mdns.Lookup(serviceName, entries)
	close(entries)
}