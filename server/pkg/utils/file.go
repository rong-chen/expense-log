package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile 将字节写入指定路径（自动创建父目录）
func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
