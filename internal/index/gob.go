package index

import (
	"bytes"
	"encoding/gob"
	"log"
)

// gobEncode gob编码通用函数
func gobEncode[T any](v T) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// gobDecode gob解码通用函数
func gobDecode[T any](data []byte, v *T) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}

// badgerLogger 简单日志实现，满足 badger.Logger 接口
type badgerLogger struct{}

func (badgerLogger) Errorf(format string, args ...interface{}) { log.Printf("[BADGER ERROR] "+format, args...) }
func (badgerLogger) Warningf(format string, args ...interface{}) { log.Printf("[BADGER WARN] "+format, args...) }
func (badgerLogger) Infof(format string, args ...interface{}) { log.Printf("[BADGER INFO] "+format, args...) }
func (badgerLogger) Debugf(format string, args ...interface{}) { log.Printf("[BADGER DEBUG] "+format, args...) }