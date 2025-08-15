package transfer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/ripplego/ripplego/internal/core"
)

// TCPTransport 提供TCP服务端与客户端下载端
// 协议：
// - 客户端 -> 服务端：GET <fileID> <offset> <size>\n
// - 服务端 -> 客户端：OK <size>\n 后续流式发送字节；或 ERR <msg>\n
// 简化：不做TLS与鉴权

type TCPTransport struct {
	Addr     string // 监听地址，示例 ":9001"
	RootDir  string // 文件根目录（用于根据 FileInfo.Path 读取文件）
	mu       sync.Mutex
	ln       net.Listener
}

func NewTCPTransport(addr, root string) *TCPTransport {
	return &TCPTransport{Addr: addr, RootDir: root}
}

func (t *TCPTransport) Serve(ctx context.Context) error {
	ln, err := net.Listen("tcp", t.Addr)
	if err != nil { return err }
	t.mu.Lock(); t.ln = ln; t.mu.Unlock()
	go func(){ <-ctx.Done(); ln.Close() }()
	for {
		conn, err := ln.Accept()
		if err != nil {
			select { case <-ctx.Done(): return nil; default: }
			return err
		}
		go t.handle(conn)
	}
}

func (t *TCPTransport) handle(conn net.Conn) {
	defer conn.Close()
	br := bufio.NewReader(conn)
	line, err := br.ReadString('\n')
	if err != nil { return }
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) != 4 || parts[0] != "GET" {
		fmt.Fprintf(conn, "ERR invalid request\n"); return
	}
	fileID := parts[1]
	offset, _ := strconv.ParseInt(parts[2], 10, 64)
	size, _ := strconv.ParseInt(parts[3], 10, 64)

	// 这里简单根据 fileID 查找本地路径：假定 fileID 为sha256(filePath:size)
	// demo实现：要求客户端提供完整绝对路径作为fileID（或外层维护映射）。
	// 生产逻辑应通过索引存储反查路径，此处简化。
	path := fileID
	if !filepath.IsAbs(path) {
		path = filepath.Join(t.RootDir, path)
	}
	f, err := os.Open(path)
	if err != nil { fmt.Fprintf(conn, "ERR %v\n", err); return }
	defer f.Close()
	if _, err := f.Seek(offset, io.SeekStart); err != nil { fmt.Fprintf(conn, "ERR %v\n", err); return }

	fmt.Fprintf(conn, "OK %d\n", size)
	if _, err := io.CopyN(conn, f, size); err != nil {
		return
	}
}

func (t *TCPTransport) Download(ctx context.Context, node core.Node, fileID core.FileID, chunk core.ChunkInfo, w io.Writer) error {
	addr := node.Address
	if addr == "" { return errors.New("empty node address") }
	conn, err := net.Dial("tcp", addr)
	if err != nil { return err }
	defer conn.Close()
	// 发送请求
	req := fmt.Sprintf("GET %s %d %d\n", fileID, chunk.Offset, chunk.Size)
	if _, err := io.WriteString(conn, req); err != nil { return err }
	// 读取响应头
	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	if err != nil { return err }
	status = strings.TrimSpace(status)
	if !strings.HasPrefix(status, "OK ") {
		return fmt.Errorf("bad response: %s", status)
	}
	// 读数据
	_, err = io.CopyN(w, br, chunk.Size)
	return err
}