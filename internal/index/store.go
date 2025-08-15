package index

import (
	"errors"
	"path/filepath"
	"sync"

	badger "github.com/dgraph-io/badger/v4"

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

// Badger 持久化实现
// 数据布局：
// - file/<fileID> -> gob(FileInfo)
// - chunks/<fileID> -> gob([]ChunkInfo)
// - nodechunks/<nodeID> -> gob(NodeChunkMap)

type BadgerStore struct {
	db *badger.DB
}

func NewBadgerStore(dir string) (*BadgerStore, error) {
	if dir == "" { dir = ".ripplego/index" }
	abs, err := filepath.Abs(dir)
	if err != nil { return nil, err }
	opts := badger.DefaultOptions(abs)
	opts = opts.WithLogger(badgerLogger{})
	db, err := badger.Open(opts)
	if err != nil { return nil, err }
	return &BadgerStore{db: db}, nil
}

func (s *BadgerStore) Close() error { return s.db.Close() }

func key(prefix, id string) []byte { return []byte(prefix + "/" + id) }

func encode[T any](v T) ([]byte, error) { return gobEncode(v) }
func decode[T any](b []byte, v *T) error { return gobDecode(b, v) }

func (s *BadgerStore) SaveFile(info core.FileInfo) error {
	b, err := encode(info)
	if err != nil { return err }
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key("file", string(info.ID)), b).WithTTL(0))
	})
}

func (s *BadgerStore) GetFile(id core.FileID) (core.FileInfo, error) {
	var out core.FileInfo
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key("file", string(id)))
		if err != nil { return err }
		return item.Value(func(val []byte) error { return decode(val, &out) })
	})
	return out, err
}

func (s *BadgerStore) ListFiles() []core.FileInfo {
	var out []core.FileInfo
	_ = s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("file/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			_ = item.Value(func(val []byte) error {
				var fi core.FileInfo
				if err := decode(val, &fi); err == nil { out = append(out, fi) }
				return nil
			})
		}
		return nil
	})
	return out
}

func (s *BadgerStore) SaveChunks(fileID core.FileID, chunks []core.ChunkInfo) error {
	b, err := encode(chunks)
	if err != nil { return err }
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key("chunks", string(fileID)), b)
	})
}

func (s *BadgerStore) GetChunks(fileID core.FileID) ([]core.ChunkInfo, error) {
	var out []core.ChunkInfo
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key("chunks", string(fileID)))
		if err != nil { return err }
		return item.Value(func(val []byte) error { return decode(val, &out) })
	})
	return out, err
}

func (s *BadgerStore) SaveNodeChunks(m core.NodeChunkMap) error {
	b, err := encode(m)
	if err != nil { return err }
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key("nodechunks", string(m.NodeID)), b)
	})
}

func (s *BadgerStore) GetNodeChunks(nodeID core.NodeID) (core.NodeChunkMap, error) {
	var out core.NodeChunkMap
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key("nodechunks", string(nodeID)))
		if err != nil { return err }
		return item.Value(func(val []byte) error { return decode(val, &out) })
	})
	return out, err
}

// gob 编解码工具与简易Logger
// 为了避免引入额外依赖，这里用标准库gob持久化结构