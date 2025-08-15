# RippleGo

一个单文件可执行的轻量级 P2P 文件分享工具。

## 目标
- 无中心服务器（去中心化）
- 支持局域网和互联网
- 多线程并发下载，速度快
- 自动节点发现（mDNS / DHT）
- 跨平台（macOS / Linux / Windows）

## 架构
CLI -> 控制模块 -> (节点发现/文件索引/分片传输/状态管理) -> 本地存储

## 快速开始
```
go build -o p2p-tool ./cmd/ripplego
./p2p-tool --help
```

## 命令
- share <file>：分享文件
- get <file_id>：下载文件
- list：查看当前可下载文件

## 开发
- Go 1.21+
- 使用 Cobra 实现 CLI
- 使用 mDNS/DHT 发现节点
- TCP + goroutine 实现分片传输