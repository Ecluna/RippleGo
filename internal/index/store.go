package index

import (
	"errors"
	"sync"

	"github.com/ripplego/ripplego/internal/core"
)

// IndexStore 定义文件元数据与分片映射的存储接口
type IndexStore interface {
	SaveFile(info core.FileInfo) error
	GetFile(id core.FileID) (core.FileInfo, error)
	ListFiles() []core.FileInfo
}

// MemoryStore 内存实现，后续可替换为持久化
type MemoryStore struct {
	mu    sync.RWMutex
	files map[core.FileID]core.FileInfo
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		files: make(map[core.FileID]core.FileInfo),
	}
}

func (s *MemoryStore) SaveFile(info core.FileInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[info.ID] = info
	return nil
}

func (s *MemoryStore) GetFile(id core.FileID) (core.FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fi, ok := s.files[id]
	if !ok {
		return core.FileInfo{}, errors.New("file not found")
	}
	return fi, nil
}

func (s *MemoryStore) ListFiles() []core.FileInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]core.FileInfo, 0, len(s.files))
	for _, v := range s.files {
		out = append(out, v)
	}
	return out
}