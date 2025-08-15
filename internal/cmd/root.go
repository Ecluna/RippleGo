package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/schollz/progressbar/v3"
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
	return &cobra.Command{
		Use:   "list",
		Short: "查看当前可下载文件 (占位)",
		RunE: func(cmd *cobra.Command, args []string) error {
			bar := progressbar.NewOptions(100,
				progressbar.OptionSetDescription("扫描网络中的共享..."),
				progressbar.OptionShowBytes(false),
			)
			for i := 0; i < 100; i++ {
				_ = bar.Add(1)
			}
			fmt.Println("\n暂未实现节点发现与索引，后续迭代补全。")
			return nil
		},
	}
}