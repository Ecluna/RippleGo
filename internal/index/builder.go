package index

import (
	"fmt"
	"math"
	"path/filepath"
	"time"

	"github.com/ripplego/ripplego/internal/core"
)

// BuildFileIndex 读取文件，计算哈希，生成文件元信息与分片列表
// 返回 FileInfo 和对应的 ChunkInfo 列表
func BuildFileIndex(filePath string, chunkSize int64) (core.FileInfo, []core.ChunkInfo, error) {
	fileHash, size, err := ComputeFileSHA256(filePath)
	if err != nil {
		return core.FileInfo{}, nil, err
	}

	if chunkSize <= 0 {
		chunkSize = 4 * 1024 * 1024 // 默认 4MB
	}
	chunkCount := int(int64(math.Ceil(float64(size) / float64(chunkSize))))

	fi := core.FileInfo{
		ID:         core.GenerateFileID(filePath, size),
		Name:       filepath.Base(filePath),
		Path:       filePath,
		Size:       size,
		Hash:       fileHash,
		ChunkSize:  chunkSize,
		ChunkCount: chunkCount,
		CreatedAt:  time.Now(),
		Description:"",
	}

	chunks := make([]core.ChunkInfo, 0, chunkCount)
	var offset int64 = 0
	for i := 0; i < chunkCount; i++ {
		remaining := size - offset
		sz := chunkSize
		if remaining < chunkSize {
			sz = remaining
		}
		cid := core.GenerateChunkID(fi.ID, i)
		chash, err := ComputeChunkSHA256(filePath, offset, sz)
		if err != nil { return core.FileInfo{}, nil, err }
		chunks = append(chunks, core.ChunkInfo{
			ID:     cid,
			FileID: fi.ID,
			Index:  i,
			Size:   sz,
			Hash:   chash,
			Offset: offset,
		})
		offset += sz
	}

	if offset != size {
		return core.FileInfo{}, nil, fmt.Errorf("size mismatch: got offset=%d, size=%d", offset, size)
	}
	return fi, chunks, nil
}