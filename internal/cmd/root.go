package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/ripplego/ripplego/internal/discovery"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ripplego",
		Short: "RippleGo: 轻量级 P2P 文件分享工具",
		Long:  `RippleGo 是一个去中心化、跨平台、支持并发下载的 P2P 文件分享工具。`,
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newShareCmd())
	cmd.AddCommand(newGetCmd())

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
	var name string

	c := &cobra.Command{
		Use:   "list",
		Short: "发现局域网节点 (UDP 广播)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()

			finder := discovery.NewUDPFinderQuery(port)
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

	c.Flags().IntVarP(&port, "port", "p", 7788, "UDP 广播端口")
	c.Flags().StringVar(&name, "name", "ripplego", "节点名称")
	return c
}

func newServeCmd() *cobra.Command {
	var port int
	var name string

	c := &cobra.Command{
		Use:   "serve",
		Short: "启动节点并广播存在 (UDP 广播)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			finder := discovery.NewUDPFinder(name, port)
			if err := finder.Start(ctx); err != nil {
				return err
			}
			fmt.Printf("RippleGo 节点已启动，正在通过 UDP:%d 广播，名称=%s。按 Ctrl+C 停止。\n", port, name)

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			<-sigCh
			fmt.Println("\n正在退出...")
			return finder.Stop()
		},
	}

	c.Flags().IntVarP(&port, "port", "p", 7788, "UDP 广播端口")
	c.Flags().StringVar(&name, "name", "ripplego", "节点名称")
	return c
}