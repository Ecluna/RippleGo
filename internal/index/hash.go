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