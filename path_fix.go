package nzgoutil

import (
	"strings"

	"github.com/NightmareZero/nzgoutil/common"
)

// 如果末尾没有斜杠，则添加斜杠
// windows  linux
func FixPathSlash(path string) string {
	if !strings.HasSuffix(path, common.PATH_DELIMITER) {
		return path + common.PATH_DELIMITER
	}
	return path
}
