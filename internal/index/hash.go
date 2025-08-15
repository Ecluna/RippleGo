package index

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// ComputeFileSHA256 计算文件的SHA-256哈希，返回十六进制字符串
func ComputeFileSHA256(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), size, nil
}

// ComputeChunkSHA256 计算指定偏移与长度的分片哈希
func ComputeChunkSHA256(path string, offset int64, size int64) (string, error) {
	f, err := os.Open(path)
	if err != nil { return "", err }
	defer f.Close()
	if _, err := f.Seek(offset, io.SeekStart); err != nil { return "", err }
	h := sha256.New()
	lr := io.LimitReader(f, size)
	if _, err := io.Copy(h, lr); err != nil { return "", err }
	return hex.EncodeToString(h.Sum(nil)), nil
}