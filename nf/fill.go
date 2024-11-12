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
		if field.OverflowInt(intValue) {
			return fmt.Errorf("整数值 %d 超出类型 %s 范围", intValue, field.Type())
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 添加对 uint 类型的支持
		uintValue, err := strconv.ParseUint(fmt.Sprint(value), 10, 64)
		if err != nil {
			return err
		}
		if field.OverflowUint(uintValue) {
			return fmt.Errorf("无符号整数值 %d 超出类型 %s 范围", uintValue, field.Type())
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(fmt.Sprint(value), 64)
		if err != nil {
			return err
		}
		if field.OverflowFloat(floatValue) {
			return fmt.Errorf("浮点数值 %f 超出类型 %s 范围", floatValue, field.Type())
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
			return setTimeField(field, value)
		}
		mapValue, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map[string]interface{} for struct, got %T", value)
		}
		return setStructField(field, mapValue)
	case reflect.Ptr:
		// 处理指针类型
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setField(field.Elem(), value)
	case reflect.Slice:
		return setSliceField(field, value)
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

// setSliceField 处理切片类型
func setSliceField(field reflect.Value, value interface{}) error {
	switch v := value.(type) {
	case []interface{}:
		slice := reflect.MakeSlice(field.Type(), len(v), len(v))
		for i := 0; i < len(v); i++ {
			if err := setField(slice.Index(i), v[i]); err != nil {
				return err
			}
		}
		field.Set(slice)
		return nil
	case string:
		// 处理单个字符串值作为切片的情况
		slice := reflect.MakeSlice(field.Type(), 1, 1)
		if err := setField(slice.Index(0), v); err != nil {
			return err
		}
		field.Set(slice)
		return nil
	default:
		return fmt.Errorf("切片字段需要 slice 类型，得到 %T", value)
	}
}

// setTimeField 处理时间类型
func setTimeField(field reflect.Value, value interface{}) error {
	switch v := value.(type) {
	case string:
		// 尝试多种时间格式
		layouts := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				field.Set(reflect.ValueOf(t))
				return nil
			}
		}
		return fmt.Errorf("无法解析时间字符串 '%s'", v)
	case time.Time:
		field.Set(reflect.ValueOf(v))
		return nil
	default:
		return fmt.Errorf("时间字段需要字符串类型，得到 %T", value)
	}
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
