package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/NightmareZero/nzgoutil/config/prop"
	"github.com/NightmareZero/nzgoutil/utilp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// 支持的配置类型
type ConfigType string

const (
	Json       ConfigType = "json"
	Yaml       ConfigType = "yaml"
	Properties ConfigType = "properties" // 以 '=' 分割的配置文件 关键字 '=' '#'
)

type Config struct {
	conf map[string]any
}

func (c *Config) Val(path string) (ret any, getted bool) {
	s := utilp.SplitBy(path, ".", "\\")
	var cursor any = c.conf
	for _, v := range s {
		parent, ok := cursor.(map[string]any)
		if !ok {
			return
		}
		cursor = parent[v]

	}
	ret = cursor
	getted = true
	return
}

func (c *Config) Str(path string) string {
	cv, ok := c.Val(path)
	if ok {
		return fmt.Sprintf("%v", cv)
	}
	return ""
}

// 解析配置文件(并放置于map中)
// $content 文件内容
// $configType 文件类型
// $target 目标数据结构
func FromFile(path string, configType ConfigType) (Config, error) {
	c := Config{map[string]any{}}
	return c, ParseFile(path, configType, c.conf)
}

// 解析配置文件(并放置于map中)
// $content 文件内容
// $configType 文件类型
// $target 目标数据结构
func From(content []byte, configType ConfigType) (Config, error) {
	c := Config{map[string]any{}}
	return c, Parse(content, configType, c.conf)
}

// 解析配置文件
// $content 文件内容
// $configType 文件类型
// $target 目标数据结构
func Parse[T any](content []byte, configType ConfigType, target T) error {
	switch configType {
	case Json:
		return json.Unmarshal(content, &target)
	case Yaml:
		return yaml.Unmarshal(content, &target)
	case Properties:
		return prop.Unmarshal(content, target)
	}
	return errors.New("Invalid config type")
}

// 解析配置文件
// $path 文件地址
// $configType 文件类型
// $target 目标数据结构
func ParseFile[T any](path string, configType ConfigType, target T) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return errors.WithStack(err)
	}

	return Parse(b, configType, target)
}
