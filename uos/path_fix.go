package uos

import (
	"strings"
)

// 如果末尾没有斜杠，则添加斜杠
// windows  linux
func FixPathSlash(path string) string {
	if !strings.HasSuffix(path, PATH_DELIMITER) {
		return path + PATH_DELIMITER
	}
	return path
}
