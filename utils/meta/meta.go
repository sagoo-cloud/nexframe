package meta

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/utils/gstructs"
	"reflect"
)

const (
	metaAttributeName = "Meta"      // metaAttributeName is the attribute name of metadata in struct.
	metaTypeName      = "meta.Meta" // metaTypeName is for type string comparison.
)

// Meta 用于API定义的元数据
type Meta struct {
	Path          string
	Method        string
	Summary       string
	Tags          string
	ExtraMetadata map[string]string
}

func InitMeta(object interface{}) error {
	v := reflect.ValueOf(object)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("object must be a pointer to struct, got %T", object)
	}

	v = v.Elem()
	metaField := v.FieldByName(metaAttributeName)
	if !metaField.IsValid() {
		return fmt.Errorf("object does not have a Meta field")
	}
	if metaField.Type().String() != metaTypeName {
		return fmt.Errorf("Meta field is of type %s, expected %s", metaField.Type().String(), metaTypeName)
	}

	metaValue := reflect.New(metaField.Type()).Elem()
	structType := v.Type()
	if field, ok := structType.FieldByName(metaAttributeName); ok {
		tags := gstructs.ParseTag(string(field.Tag))

		// 检查必要的标签是否存在
		requiredTags := []string{"path", "method"}
		for _, tag := range requiredTags {
			if _, exists := tags[tag]; !exists {
				return fmt.Errorf("required tag '%s' is missing in Meta field", tag)
			}
		}

		metaValue.FieldByName("Path").SetString(tags["path"])
		metaValue.FieldByName("Method").SetString(tags["method"])
		metaValue.FieldByName("Summary").SetString(tags["summary"])
		metaValue.FieldByName("Tags").SetString(tags["tags"])

		extraMetadata := make(map[string]string)
		for k, v := range tags {
			if k != "path" && k != "method" && k != "summary" && k != "tags" {
				extraMetadata[k] = v
			}
		}

		// 处理结构体中的其他字段的标签
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			if field.Name != metaAttributeName {
				fieldTags := gstructs.ParseTag(string(field.Tag))
				for k, v := range fieldTags {
					if k != "json" && k != "v" { // 排除常见的标签
						extraMetadata[k] = v
					}
				}
			}
		}

		metaValue.FieldByName("ExtraMetadata").Set(reflect.ValueOf(extraMetadata))
	} else {
		return fmt.Errorf("Meta field not found in struct")
	}

	metaField.Set(metaValue)
	return nil
}
func Data(object interface{}) map[string]string {
	if err := InitMeta(object); err != nil {
		return nil
	}
	v := reflect.ValueOf(object)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	metaField := v.FieldByName(metaAttributeName)
	if metaField.IsValid() && metaField.Type().String() == metaTypeName {
		result := make(map[string]string)
		result["path"] = metaField.FieldByName("Path").String()
		result["method"] = metaField.FieldByName("Method").String()
		result["summary"] = metaField.FieldByName("Summary").String()
		result["tags"] = metaField.FieldByName("Tags").String()
		extraMetadata := metaField.FieldByName("ExtraMetadata").Interface().(map[string]string)
		for k, v := range extraMetadata {
			result[k] = v
		}
		return result
	}
	return map[string]string{}
}
