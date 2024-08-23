package nf

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/nf/g"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (f *APIFramework) fillStruct(data map[string]interface{}, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			// 处理嵌入字段
			if err := f.fillStruct(data, v.Field(i)); err != nil {
				return err
			}
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		value, exists := data[jsonTag]
		if !exists {
			continue
		}

		if err := setField(v.Field(i), value); err != nil {
			return fmt.Errorf("error setting field %s: %v", field.Name, err)
		}
	}

	return nil
}
func (f *APIFramework) fillSlice(data []interface{}, v reflect.Value) error {
	slice := reflect.MakeSlice(v.Type(), len(data), len(data))
	for i, item := range data {
		if err := setField(slice.Index(i), item); err != nil {
			return err
		}
	}
	v.Set(slice)
	return nil
}

func (f *APIFramework) fillMap(data map[string]interface{}, v reflect.Value) error {
	m := reflect.MakeMap(v.Type())
	for k, val := range data {
		key := reflect.New(v.Type().Key()).Elem()
		if err := setField(key, k); err != nil {
			return err
		}
		value := reflect.New(v.Type().Elem()).Elem()
		if err := setField(value, val); err != nil {
			return err
		}
		m.SetMapIndex(key, value)
	}
	v.Set(m)
	return nil
}

func setIntField(field reflect.Value, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetInt(int64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		field.SetInt(int64(v.Float()))
	case reflect.String:
		intValue, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	default:
		return fmt.Errorf("cannot set int field from type %s", v.Type())
	}
	return nil
}

func setUintField(field reflect.Value, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetUint(uint64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(v.Uint())
	case reflect.Float32, reflect.Float64:
		field.SetUint(uint64(v.Float()))
	case reflect.String:
		uintValue, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	default:
		return fmt.Errorf("cannot set uint field from type %s", v.Type())
	}
	return nil
}

func setFloatField(field reflect.Value, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetFloat(float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetFloat(float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		field.SetFloat(v.Float())
	case reflect.String:
		floatValue, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("cannot set float field from type %s", v.Type())
	}
	return nil
}

func setBoolField(field reflect.Value, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		field.SetBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetBool(v.Int() != 0)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetBool(v.Uint() != 0)
	case reflect.String:
		boolValue, err := strconv.ParseBool(v.String())
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	default:
		return fmt.Errorf("cannot set bool field from type %s", v.Type())
	}
	return nil
}

// 可以保持不变或根据需要进行类似的调整
func setField(field reflect.Value, value interface{}) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprint(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(fmt.Sprint(value), 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(fmt.Sprint(value))
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			timeStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("expected string for time, got %T", value)
			}
			t, err := time.Parse(time.RFC3339, timeStr)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(t))
		}
	case reflect.Map:
		mapValue, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map for field, got %T", value)
		}
		newMap := reflect.MakeMap(field.Type())
		for k, v := range mapValue {
			newMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
		}
		field.Set(newMap)
	case reflect.Interface:
		field.Set(reflect.ValueOf(value))
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}
	return nil
}

func setTimeField(field reflect.Value, value interface{}) error {
	switch v := value.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("error parsing time string: %v", err)
		}
		field.Set(reflect.ValueOf(t))
	case float64:
		field.Set(reflect.ValueOf(time.Unix(int64(v), 0)))
	case int64:
		field.Set(reflect.ValueOf(time.Unix(v, 0)))
	default:
		return fmt.Errorf("unsupported type for time field: %T", value)
	}
	return nil
}

func setStructField(field reflect.Value, data map[string]interface{}) error {
	for i := 0; i < field.NumField(); i++ {
		structField := field.Type().Field(i)

		if structField.Anonymous {
			if err := setAnonymousField(field.Field(i), data); err != nil {
				return err
			}
			continue
		}

		jsonTag := structField.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = structField.Name
		}
		jsonTag = strings.Split(jsonTag, ",")[0] // 处理 omitempty 等选项

		if jsonTag == "-" {
			continue
		}

		fieldValue := field.Field(i)
		if value, ok := data[jsonTag]; ok {
			if err := setField(fieldValue, value); err != nil {
				return fmt.Errorf("error setting field %s: %v", structField.Name, err)
			}
		}
	}
	return nil
}

func setAnonymousField(field reflect.Value, data map[string]interface{}) error {
	switch field.Interface().(type) {
	case g.Meta:
		// 处理 g.Meta 字段
		meta := field.Addr().Interface().(*g.Meta)
		if path, ok := data["Path"].(string); ok {
			meta.Path = path
		}
		if method, ok := data["Method"].(string); ok {
			meta.Method = method
		}
		if summary, ok := data["Summary"].(string); ok {
			meta.Summary = summary
		}
		if tags, ok := data["Tags"].(string); ok {
			meta.Tags = tags
		}
		if extraMetadata, ok := data["ExtraMetadata"].(map[string]interface{}); ok {
			meta.ExtraMetadata = make(map[string]string)
			for k, v := range extraMetadata {
				if strValue, ok := v.(string); ok {
					meta.ExtraMetadata[k] = strValue
				} else {
					meta.ExtraMetadata[k] = fmt.Sprint(v)
				}
			}
		}
	default:
		// 对于其他类型的匿名字段，递归处理
		return setStructField(field, data)
	}
	return nil
}

func setSliceField(field reflect.Value, value interface{}) error {
	slice, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected slice, got %T", value)
	}
	sliceValue := reflect.MakeSlice(field.Type(), len(slice), len(slice))
	for i, item := range slice {
		if err := setField(sliceValue.Index(i), item); err != nil {
			return err
		}
	}
	field.Set(sliceValue)
	return nil
}

func setMapField(field reflect.Value, v reflect.Value) error {
	if v.Kind() != reflect.Map {
		return fmt.Errorf("expected map, got %v", v.Kind())
	}

	mapValue := reflect.MakeMap(field.Type())
	for _, key := range v.MapKeys() {
		mapValue.SetMapIndex(key, v.MapIndex(key))
	}
	field.Set(mapValue)
	return nil
}

func setPtrField(field reflect.Value, value interface{}) error {
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}
	return setField(field.Elem(), value)
}
