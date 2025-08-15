package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/ripplego/ripplego/internal/discovery"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "p2p-tool",
		Short: "RippleGo: 轻量级 P2P 文件分享工具",
		Long:  `RippleGo 是一个去中心化、跨平台、支持并发下载的 P2P 文件分享工具。`,
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("RippleGo v0.1.0")
		},
	}
}

func newListCmd() *cobra.Command {
	var port int
	var host string

	c := &cobra.Command{
		Use:   "list",
		Short: "查看当前可下载文件 (发现局域网节点)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()

			finder := discovery.NewMDNSFinder(host, port)
			if err := finder.Start(ctx); err != nil {
				return err
			}
			time.Sleep(2 * time.Second)
			nodes := finder.Nodes()

			bar := progressbar.NewOptions(100,
				progressbar.OptionSetDescription("扫描网络中的节点..."),
				progressbar.OptionShowBytes(false),
			)
			for i := 0; i < 100; i++ {
				_ = bar.Add(1)
			}
			fmt.Printf("\n发现节点数: %d\n", len(nodes))
			for _, n := range nodes {
				fmt.Printf("- %s (%s)\n", n.ID, n.Address)
			}
			return finder.Stop()
		},
	}

	c.Flags().IntVarP(&port, "port", "p", 7788, "mDNS 广播端口")
	c.Flags().StringVar(&host, "host", "ripplego.local", "mDNS 主机名")
	return c
}