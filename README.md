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
go build -o ripplego ./cmd/ripplego
./ripplego --help
```

## 命令
- 分享文件（生成并持久化索引）
  ```bash
  ripplego share -f /path/to/file --chunk-size 4194304 --store .ripplego/index
  ```
  - 关键参数：
    - -f, --file：要分享的文件路径
    - --chunk-size：分片大小（字节），默认 4MB（4194304）
    - --store：索引持久化目录，默认 .ripplego/index

- 下载文件（并发/断点续传-简化）
  ```bash
  ripplego get --file-id <FILE_ID> --addr 127.0.0.1:9001 --out /path/to/output --store .ripplego/index --workers 4
  ```
  - 关键参数：
    - --file-id：目标文件 ID（由 share 命令输出）
    - --addr：源节点地址（示例：127.0.0.1:9001）
    - --out：输出文件路径（默认使用文件名）
    - --store：索引持久化目录
    - --workers：并发下载的工作协程数，默认 4

- 发现局域网节点（UDP 广播）
  ```bash
  ripplego list --port 7788 --name ripplego
  ```
  - 关键参数：
    - --port：UDP 广播端口，默认 7788
    - --name：节点名称，默认 ripplego

## 开发
- Go 1.21+
- 使用 Cobra 实现 CLI
- 使用 mDNS/DHT 发现节点
- TCP + goroutine 实现分片传输