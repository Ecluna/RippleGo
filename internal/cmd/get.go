package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/ripplego/ripplego/internal/core"
	"github.com/ripplego/ripplego/internal/index"
	"github.com/ripplego/ripplego/internal/transfer"
)

func newGetCmd() *cobra.Command {
	var (
		fileID   string
		outPath  string
		addr     string
		storeDir string
		workers  int
	)

	c := &cobra.Command{
		Use:   "get",
		Short: "从远端节点下载文件(并发/断点续传-简化)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fileID == "" || addr == "" {
				return fmt.Errorf("请提供 --file-id 与 --addr")
			}

			bs, err := index.NewBadgerStore(storeDir)
			if err != nil { return err }
			defer bs.Close()

			fi, err := bs.GetFile(core.FileID(fileID))
			if err != nil { return err }
			chunks, err := bs.GetChunks(fi.ID)
			if err != nil { return err }

			if outPath == "" { outPath = filepath.Base(fi.Name) }
			tmpPath := outPath + ".part"
			f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil { return err }
			defer f.Close()
			if stat, _ := f.Stat(); stat.Size() < fi.Size {
				_ = f.Truncate(fi.Size)
			}

			bar := progressbar.DefaultBytes(fi.Size, "downloading")
			tr := transfer.NewTCPTransport("", "")
			node := core.Node{Address: addr}

			sem := make(chan struct{}, workers)
			var wg sync.WaitGroup
			var firstErr error
			var mu sync.Mutex
			for _, ch := range chunks {
				ch := ch
				sem <- struct{}{}
				wg.Add(1)
				go func(){
					defer func(){ <-sem; wg.Done() }()
					buf := bytes.NewBuffer(make([]byte, 0, ch.Size))
					if err := tr.Download(cmd.Context(), node, fi.ID, ch, buf); err != nil {
						mu.Lock(); if firstErr == nil { firstErr = err }; mu.Unlock(); return
					}
					// 写入目标文件指定偏移
					mu.Lock()
					if _, err := f.WriteAt(buf.Bytes(), ch.Offset); err != nil && firstErr==nil { firstErr = err }
					_ = bar.Add64(ch.Size)
					mu.Unlock()
				}()
			}
			wg.Wait()
			if firstErr != nil { return firstErr }
			_ = f.Close()
			return os.Rename(tmpPath, outPath)
		},
	}

	c.Flags().StringVar(&fileID, "file-id", "", "目标文件ID")
	c.Flags().StringVar(&outPath, "out", "", "输出文件路径")
	c.Flags().StringVar(&addr, "addr", "", "源节点地址，例如 127.0.0.1:9001")
	c.Flags().StringVar(&storeDir, "store", ".ripplego/index", "索引持久化目录")
	c.Flags().IntVar(&workers, "workers", 4, "并发下载的工作协程数")
	return c
}