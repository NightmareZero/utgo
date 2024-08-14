package prop

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/NightmareZero/nzgoutil/config/node"
	"github.com/NightmareZero/nzgoutil/utilp"
)

const TAG_NAME = "prop"

type propertiesReader struct {
	cache map[string]node.Node
}

func Unmarshal(data []byte, v any) error {
	pr := propertiesReader{}
	pr.init()

	// read
	s := bufio.NewScanner(bytes.NewReader(data))
	s.Split(bufio.ScanLines)

	for s.Scan() {
		pr.unmarshalLine(s.Text())
	}

	//parse

	panic("unimplemented")
}

func (r *propertiesReader) init() {
	r.cache = map[string]node.Node{}
}

func (r *propertiesReader) unmarshalLine(line string) {
	// 移除前后空格
	l := strings.TrimSpace(line)

	// 如果是注释，则跳过本行
	if strings.HasPrefix(l, "#") {
		return
	}

	// 获取key
	sh := utilp.SplitHead(l, "=", "\\")
	// 拆分后方注释获取val
	sv := utilp.SplitHead(sh[1], "#", "\\")

	nodeKey := strings.TrimSpace(sh[0])
	nodeVal := strings.TrimSpace(sv[0])

	r.cache[nodeKey] = nodeVal
}
