package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ripplego/ripplego/internal/index"
)

func newShareCmd() *cobra.Command {
	var (
		filePath  string
		chunkSize int64
	)

	c := &cobra.Command{
		Use:   "share",
		Short: "分享本地文件，生成索引",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("请使用 --file 指定要分享的文件路径")
			}

			fi, chunks, err := index.BuildFileIndex(filePath, chunkSize)
			if err != nil {
				return err
			}

			store := index.NewMemoryStore()
			if err := store.SaveFile(fi); err != nil { return err }
			if err := store.SaveChunks(fi.ID, chunks); err != nil { return err }

			// 这里暂存于内存，后续可以扩展为持久化到本地DB
			fmt.Printf("已建立索引：%s\n- 文件ID: %s\n- 大小: %d bytes\n- 分片: %d 个 (chunkSize=%d)\n",
				fi.Name, fi.ID, fi.Size, fi.ChunkCount, fi.ChunkSize)

			_ = context.TODO() // 预留后续启动传输服务等
			_ = time.Second
			return nil
		},
	}

	c.Flags().StringVarP(&filePath, "file", "f", "", "要分享的文件路径")
	c.Flags().Int64Var(&chunkSize, "chunk-size", 4*1024*1024, "分片大小（字节），默认4MB")
	return c
}