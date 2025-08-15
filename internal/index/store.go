package index

import (
	"errors"
	"sync"

	"github.com/ripplego/ripplego/internal/core"
)

// IndexStore 定义文件元数据与分片映射的存储接口
// 后续可替换为持久化实现（如Badger/SQLite等）
type IndexStore interface {
	SaveFile(info core.FileInfo) error
	GetFile(id core.FileID) (core.FileInfo, error)
	ListFiles() []core.FileInfo

	SaveChunks(fileID core.FileID, chunks []core.ChunkInfo) error
	GetChunks(fileID core.FileID) ([]core.ChunkInfo, error)

	SaveNodeChunks(m core.NodeChunkMap) error
	GetNodeChunks(nodeID core.NodeID) (core.NodeChunkMap, error)
}

// MemoryStore 内存实现，后续可替换为持久化
type MemoryStore struct {
	mu         sync.RWMutex
	files      map[core.FileID]core.FileInfo
	chunks     map[core.FileID][]core.ChunkInfo
	nodeChunks map[core.NodeID]core.NodeChunkMap
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		files:      make(map[core.FileID]core.FileInfo),
		chunks:     make(map[core.FileID][]core.ChunkInfo),
		nodeChunks: make(map[core.NodeID]core.NodeChunkMap),
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

func (s *MemoryStore) SaveChunks(fileID core.FileID, chunks []core.ChunkInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chunks[fileID] = append([]core.ChunkInfo(nil), chunks...)
	return nil
}

func (s *MemoryStore) GetChunks(fileID core.FileID) ([]core.ChunkInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	chs, ok := s.chunks[fileID]
	if !ok {
		return nil, errors.New("chunks not found")
	}
	out := append([]core.ChunkInfo(nil), chs...)
	return out, nil
}

func (s *MemoryStore) SaveNodeChunks(m core.NodeChunkMap) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeChunks[m.NodeID] = m
	return nil
}

func (s *MemoryStore) GetNodeChunks(nodeID core.NodeID) (core.NodeChunkMap, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.nodeChunks[nodeID]
	if !ok {
		return core.NodeChunkMap{}, errors.New("node chunk map not found")
	}
	return m, nil
}