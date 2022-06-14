package upath

import (
	"strings"

	"github.com/NightmareZero/m-go-starter/common/uconst"
)

// 如果末尾没有斜杠，则添加斜杠
// windows  linux
func FixPathSlash(path string) string {
	if !strings.HasSuffix(path, uconst.PATH_DELIMITER) {
		return path + uconst.PATH_DELIMITER
	}
	return path
}
