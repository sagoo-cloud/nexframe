package convert

import (
	"reflect"
	"strings"
)

func Map(value interface{}, tags ...string) map[string]interface{} {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	t := v.Type()

	// If it's a pointer, get the underlying element
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	switch t.Kind() {
	case reflect.Map:
		return convertMap(v)
	case reflect.Struct:
		return convertStruct(v, tags)
	default:
		return nil
	}
}

func convertMap(v reflect.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for _, key := range v.MapKeys() {
		result[key.String()] = v.MapIndex(key).Interface()
	}
	return result
}

func convertStruct(v reflect.Value, tags []string) map[string]interface{} {
	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		var key string
		for _, tag := range append(tags, "c", "gconv", "json") {
			if tagValue := field.Tag.Get(tag); tagValue != "" {
				key = strings.Split(tagValue, ",")[0]
				break
			}
		}

		if key == "" {
			key = field.Name
		}

		if key == "-" {
			continue
		}

		result[key] = value.Interface()
	}

	return result
}
