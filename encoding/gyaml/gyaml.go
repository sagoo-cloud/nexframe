// Package gyaml provides accessing and converting for YAML content.
package gyaml

import (
	"bytes"
	"encoding/json"
	"github.com/sagoo-cloud/nexframe/errors/gerror"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strings"

	"gopkg.in/yaml.v3"
)

// Encode encodes `value` to an YAML format content as bytes.
func Encode(value interface{}) (out []byte, err error) {
	if out, err = yaml.Marshal(value); err != nil {
		err = gerror.Wrap(err, `yaml.Marshal failed`)
	}
	return
}

// EncodeIndent encodes `value` to an YAML format content with indent as bytes.
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

// Decode parses `content` into and returns as map.
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

// DecodeTo parses `content` into `result`.
func DecodeTo(value []byte, result interface{}) (err error) {
	err = yaml.Unmarshal(value, result)
	if err != nil {
		err = gerror.Wrap(err, `yaml.Unmarshal failed`)
	}
	return
}

// ToJson converts `content` to JSON format content.
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
