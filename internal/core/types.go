package core

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strconv"
	"time"
)

// NodeID 节点的唯一标识符
type NodeID string

// FileID 文件的唯一标识符（SHA-256哈希）
type FileID string

// ChunkID 分片的唯一标识符
type ChunkID string

// Node 表示P2P网络中的一个节点
type Node struct {
	ID          NodeID    `json:"id"`           // 节点ID
	Address     string    `json:"address"`      // IP:Port（服务端口）
	ServicePort int       `json:"servicePort"`  // 服务端口（TCP传输）
	LastSeen    time.Time `json:"lastSeen"`     // 上次心跳时间
	Status      string    `json:"status"`       // online/offline
}

// FileInfo 文件的元数据信息
type FileInfo struct {
	ID          FileID    `json:"id"`          // 文件唯一标识
	Name        string    `json:"name"`        // 文件名
	Path        string    `json:"path"`        // 本地文件路径
	Size        int64     `json:"size"`        // 文件大小（字节）
	Hash        string    `json:"hash"`        // SHA-256哈希值
	ChunkSize   int64     `json:"chunkSize"`   // 分片大小
	ChunkCount  int       `json:"chunkCount"`  // 分片数量
	CreatedAt   time.Time `json:"createdAt"`   // 创建时间
	Description string    `json:"description"` // 文件描述
}

// ChunkInfo 分片信息
type ChunkInfo struct {
	ID     ChunkID `json:"id"`     // 分片唯一标识
	FileID FileID  `json:"fileId"` // 所属文件ID
	Index  int     `json:"index"`  // 分片索引
	Size   int64   `json:"size"`   // 分片大小
	Hash   string  `json:"hash"`   // 分片哈希值
	Offset int64   `json:"offset"` // 在文件中的偏移量
}

// NodeChunkMap 节点-分片映射表
type NodeChunkMap struct {
	NodeID   NodeID    `json:"nodeId"`   // 节点ID
	ChunkIDs []ChunkID `json:"chunkIds"` // 该节点拥有的分片列表
}

// DownloadTask 下载任务
type DownloadTask struct {
	FileID      FileID    `json:"fileId"`      // 文件ID
	ChunkID     ChunkID   `json:"chunkId"`     // 分片ID
	SourceNode  NodeID    `json:"sourceNode"`  // 源节点
	Status      string    `json:"status"`      // downloading/completed/failed
	Progress    int64     `json:"progress"`    // 下载进度（字节）
	StartTime   time.Time `json:"startTime"`   // 开始时间
	CompletedAt time.Time `json:"completedAt"` // 完成时间
}

// GenerateNodeID 生成节点ID
func GenerateNodeID(addr net.Addr) NodeID {
	hash := sha256.Sum256([]byte(addr.String() + time.Now().String()))
	return NodeID(hex.EncodeToString(hash[:])[:16])
}

// GenerateFileID 生成文件ID
func GenerateFileID(filePath string, size int64) FileID {
	data := filePath + ":" + strconv.FormatInt(size, 10)
	hash := sha256.Sum256([]byte(data))
	return FileID(hex.EncodeToString(hash[:]))
}

// GenerateChunkID 生成分片ID
func GenerateChunkID(fileID FileID, index int) ChunkID {
	data := string(fileID) + ":" + strconv.Itoa(index)
	hash := sha256.Sum256([]byte(data))
	return ChunkID(hex.EncodeToString(hash[:])[:16])
}