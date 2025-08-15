package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"syscall"
	"sync"
	"time"

	"github.com/ripplego/ripplego/internal/core"
)

// UDPFinder 基于UDP广播的局域网节点发现
// 简单有效，避免复杂依赖
type UDPFinder struct {
	port      int
	name      string
	selfID    string
	queryOnly bool
	mu        sync.RWMutex
	nodes     map[core.NodeID]core.Node
	ctx       context.Context
	cancel    context.CancelFunc
	conn      *net.UDPConn
	stopCh    chan struct{}
}

type BroadcastMsg struct {
	Type    string `json:"type"` // announce | query
	NodeID  string `json:"nodeId"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Msg     string `json:"msg"`
}

func NewUDPFinder(name string, port int) *UDPFinder {
	return &UDPFinder{
		port:   port,
		name:   name,
		selfID: fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
		nodes:  make(map[core.NodeID]core.Node),
		stopCh: make(chan struct{}),
	}
}

// NewUDPFinderQuery 创建仅用于一次性扫描的Finder，不占用固定端口
func NewUDPFinderQuery(port int) *UDPFinder {
	return &UDPFinder{
		port:      port,
		name:      "scanner",
		selfID:    fmt.Sprintf("scanner-%d", time.Now().UnixNano()),
		nodes:     make(map[core.NodeID]core.Node),
		stopCh:    make(chan struct{}),
		queryOnly: true,
	}
}

func (u *UDPFinder) Start(ctx context.Context) error {
	u.ctx, u.cancel = context.WithCancel(ctx)

	if u.queryOnly {
		// 仅扫描：使用随机端口接收回复，同时复用该 socket 发送查询
		pc, err := net.ListenPacket("udp4", ":0")
		if err != nil {
			return err
		}
		u.conn = pc.(*net.UDPConn)
		// 允许广播
		if rc, err := u.conn.SyscallConn(); err == nil {
			_ = rc.Control(func(fd uintptr) {
				_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
			})
		}
		go u.listenBroadcast()
		go u.sendQuery()
		return nil
	}

	// 服务模式：监听固定端口，启用端口复用
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var ctrlErr error
			if err := c.Control(func(fd uintptr) {
				_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
			}); err != nil {
				ctrlErr = err
			}
			return ctrlErr
		},
	}
	pc, err := lc.ListenPacket(u.ctx, "udp4", fmt.Sprintf(":%d", u.port))
	if err != nil {
		return err
	}
	u.conn = pc.(*net.UDPConn)

	go u.listenBroadcast()
	go u.sendBroadcast()
	return nil
}

func (u *UDPFinder) Stop() error {
	close(u.stopCh)
	if u.conn != nil {
		u.conn.Close()
	}
	if u.cancel != nil {
		u.cancel()
	}
	return nil
}

func (u *UDPFinder) Nodes() []core.Node {
	u.mu.Lock()
	defer u.mu.Unlock()

	// 清理超时节点（30秒未响应）
	cutoff := time.Now().Add(-30 * time.Second)
	for id, node := range u.nodes {
		if node.LastSeen.Before(cutoff) {
			delete(u.nodes, id)
		}
	}

	out := make([]core.Node, 0, len(u.nodes))
	for _, n := range u.nodes {
		out = append(out, n)
	}
	return out
}

func (u *UDPFinder) listenBroadcast() {
	buf := make([]byte, 2048)
	for {
		select {
		case <-u.stopCh:
			return
		default:
		}

		u.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := u.conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			continue
		}

		var msg BroadcastMsg
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			continue
		}

		if msg.NodeID == u.selfID {
			continue
		}

		// 若收到查询且处于服务模式，单播回复公告
		if msg.Type == "query" && !u.queryOnly {
			resp := BroadcastMsg{
				Type:    "announce",
				NodeID:  u.selfID,
				Name:    u.name,
				Address: "",
				Msg:     "RippleGo discovery",
			}
			data, _ := json.Marshal(resp)
			u.conn.WriteToUDP(data, addr)
			continue
		}

		if msg.Type != "announce" {
			continue
		}

		node := core.Node{
			ID:          core.NodeID(msg.NodeID),
			Address:     fmt.Sprintf("%s:%d", addr.IP.String(), u.port),
			ServicePort: u.port,
			LastSeen:    time.Now(),
			Status:      "online",
		}
		u.mu.Lock()
		u.nodes[node.ID] = node
		u.mu.Unlock()
	}
}


func (u *UDPFinder) interfaceBroadcastIPs() []net.IP {
	ips := []net.IP{}
	ifs, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifs {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok || ipNet.IP == nil || ipNet.IP.To4() == nil || ipNet.Mask == nil {
				continue
			}
			ip := ipNet.IP.To4()
			mask := ipNet.Mask
			bcast := net.IPv4(ip[0]|^mask[0], ip[1]|^mask[1], ip[2]|^mask[2], ip[3]|^mask[3])
			ips = append(ips, bcast)
		}
	}
	return ips
}

func (u *UDPFinder) sendBroadcast() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 单独的发送 socket，开启广播权限
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var ctrlErr error
			if err := c.Control(func(fd uintptr) {
				_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
			}); err != nil {
				ctrlErr = err
			}
			return ctrlErr
		},
	}
	pc, err := lc.ListenPacket(u.ctx, "udp4", ":0")
	if err != nil {
		return
	}
	conn := pc.(*net.UDPConn)
	defer conn.Close()

	msg := BroadcastMsg{
		Type:    "announce",
		NodeID:  u.selfID,
		Name:    u.name,
		Address: "",
		Msg:     "RippleGo discovery",
	}

	dests := []net.UDPAddr{
		{IP: net.ParseIP("127.0.0.1"), Port: u.port},
		{IP: net.IPv4bcast, Port: u.port},
	}
	for _, ip := range u.interfaceBroadcastIPs() {
		dests = append(dests, net.UDPAddr{IP: ip, Port: u.port})
	}

	for {
		select {
		case <-u.stopCh:
			return
		case <-ticker.C:
			data, _ := json.Marshal(msg)
			for _, d := range dests {
				_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
				_, _ = conn.WriteToUDP(data, &d)
			}
		}
	}
}

func (u *UDPFinder) sendQuery() {
	msg := BroadcastMsg{
		Type:    "query",
		NodeID:  u.selfID,
		Name:    u.name,
		Address: "",
		Msg:     "RippleGo discovery query",
	}

	dests := []net.UDPAddr{
		{IP: net.ParseIP("127.0.0.1"), Port: u.port},
		{IP: net.IPv4bcast, Port: u.port},
	}
	for _, ip := range u.interfaceBroadcastIPs() {
		dests = append(dests, net.UDPAddr{IP: ip, Port: u.port})
	}

	for i := 0; i < 3; i++ { // 重发几次提升成功率
		data, _ := json.Marshal(msg)
		for _, d := range dests {
			_ = u.conn.SetWriteDeadline(time.Now().Add(300 * time.Millisecond))
			_, _ = u.conn.WriteToUDP(data, &d)
		}
		time.Sleep(300 * time.Millisecond)
	}
}