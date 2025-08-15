package cmd

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"

	"github.com/ripplego/ripplego/internal/core"
	"github.com/ripplego/ripplego/internal/index"
)

func newShareCmd() *cobra.Command {
	var (
		filePath  string
		chunkSize int64
		storeDir  string
	)

	c := &cobra.Command{
		Use:   "share",
		Short: "分享本地文件，生成索引(持久化)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("请使用 --file 指定要分享的文件路径")
			}

			fi, chunks, err := index.BuildFileIndex(filePath, chunkSize)
			if err != nil {
				return err
			}

			bs, err := index.NewBadgerStore(storeDir)
			if err != nil { return err }
			defer bs.Close()

			if err := bs.SaveFile(fi); err != nil { return err }
			if err := bs.SaveChunks(fi.ID, chunks); err != nil { return err }

			// 记录本地节点持有的分片映射（简化：使用本机监听地址生成NodeID）
			addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
			nodeID := core.GenerateNodeID(addr)
			owned := make([]core.ChunkID, 0, len(chunks))
			for _, ch := range chunks { owned = append(owned, ch.ID) }
			if err := bs.SaveNodeChunks(core.NodeChunkMap{NodeID: nodeID, ChunkIDs: owned}); err != nil { return err }

			fmt.Printf("已建立并持久化索引：%s\n- 文件ID: %s\n- 大小: %d bytes\n- 分片: %d 个 (chunkSize=%d)\n",
				fi.Name, fi.ID, fi.Size, fi.ChunkCount, fi.ChunkSize)

			_ = context.TODO()
			_ = time.Second
			return nil
		},
	}

	c.Flags().StringVarP(&filePath, "file", "f", "", "要分享的文件路径")
	c.Flags().Int64Var(&chunkSize, "chunk-size", 4*1024*1024, "分片大小（字节），默认4MB")
	c.Flags().StringVar(&storeDir, "store", ".ripplego/index", "索引持久化目录")
	return c
}