package transfer

import (
	"context"
	"io"

	"github.com/ripplego/ripplego/internal/core"
)

// Transport 抽象传输接口，默认TCP实现
type Transport interface {
	Serve(ctx context.Context) error                  // 启动服务以共享本地分片
	Download(ctx context.Context, node core.Node, fileID core.FileID, chunk core.ChunkInfo, w io.Writer) error
}