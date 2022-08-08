package uos

import (
	"strings"
)

// 如果末尾没有斜杠，则添加斜杠
// windows  linux
func FixPathEndSlash(path string) string {
	if !strings.HasSuffix(path, PATH_DELIMITER) {
		return path + PATH_DELIMITER
	}
	return path
}

// 如果末尾有斜杠，则移除斜杠
// windows  linux
func FixPathNoEndSlash(path string) string {
	if !strings.HasSuffix(path, PATH_DELIMITER) {
		return path
	}
	return path[:len(path)-1]
}
