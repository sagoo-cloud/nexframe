// Package gyaml 提供YAML内容的访问和转换功能。
package gyaml

import (
	"bytes"
	"encoding/json"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"strings"

	"gopkg.in/yaml.v3"
)

// Encode 将值编码为YAML格式的字节切片。
// 内部使用yaml.Marshal进行编码，如果编码过程中发生错误，将返回包装后的错误。
func Encode(value interface{}) (out []byte, err error) {
	if out, err = yaml.Marshal(value); err != nil {
		err = gerror.Wrap(err, `yaml.Marshal failed`)
	}
	return
}

// EncodeIndent 将值编码为带有缩进的YAML格式字节切片。
// indent参数指定每行的缩进字符串。如果indent为空，则不进行缩进处理。
func EncodeIndent(value interface{}, indent string) (out []byte, err error) {
	out, err = Encode(value)
	if err != nil {
		return
	}
	if indent != "" {
		var (
			buffer = bytes.NewBuffer(nil)
			array  = strings.Split(strings.TrimSpace(string(out)), "\n")
		)
		for _, v := range array {
			buffer.WriteString(indent)
			buffer.WriteString(v)
			buffer.WriteString("\n")
		}
		out = buffer.Bytes()
	}
	return
}

// Decode 将YAML格式的内容解析为map[string]interface{}。
// 解析结果会通过convert.Map进行类型转换处理。
// 如果解析过程中发生错误，将返回包装后的错误。
func Decode(content []byte) (map[string]interface{}, error) {
	var (
		result map[string]interface{}
		err    error
	)
	if err = yaml.Unmarshal(content, &result); err != nil {
		err = gerror.Wrap(err, `yaml.Unmarshal failed`)
		return nil, err
	}
	return convert.Map(result), nil
}

// DecodeTo 将YAML格式的内容解析到指定的结构体中。
// 如果解析过程中发生错误，将返回包装后的错误。
func DecodeTo(value []byte, result interface{}) (err error) {
	err = yaml.Unmarshal(value, result)
	if err != nil {
		err = gerror.Wrap(err, `yaml.Unmarshal failed`)
	}
	return
}

// ToJson 将YAML格式的内容转换为JSON格式。
// 首先解码YAML内容，然后使用json.Marshal将结果编码为JSON格式。
func ToJson(content []byte) (out []byte, err error) {
	var (
		result interface{}
	)
	if result, err = Decode(content); err != nil {
		return nil, err
	} else {
		return json.Marshal(result)
	}
}
