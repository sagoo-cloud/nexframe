package nf

import (
	"fmt"
	"github.com/ServiceWeaver/weaver"
	"reflect"
)

// WeaverContext 包含 ServiceWeaver 相关的上下文
type WeaverContext struct {
	Services map[string]interface{} // 存储任意 ServiceWeaver 服务，包括配置
}

// AddWeaverService 自动添加 ServiceWeaver 服务
func (f *APIFramework) AddWeaverService(s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, got %T", s)
	}

	structVal := val.Elem()
	structType := structVal.Type()

	// 检查是否包含 weaver.Implements[weaver.Main]，但不要求必须实现
	_, implementsMain := s.(weaver.Implements[weaver.Main])

	// 遍历结构体字段
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// 跳过 weaver.Implements[weaver.Main] 字段
		if fieldType.Type.String() == "weaver.Implements[weaver.Main]" {
			continue
		}

		// 只处理可导出的字段
		if fieldType.PkgPath == "" {
			// 注册服务
			if f.weaverServices == nil {
				f.weaverServices = make(map[string]interface{})
			}
			f.weaverServices[fieldType.Name] = field.Interface()

			if f.debug {
				fmt.Printf("Registered ServiceWeaver service: %s of type %v\n", fieldType.Name, fieldType.Type)
			}
		} else if f.debug {
			fmt.Printf("Skipping unexported field: %s\n", fieldType.Name)
		}
	}

	if f.debug {
		if implementsMain {
			fmt.Printf("Registered ServiceWeaver main component: %T\n", s)
		} else {
			fmt.Printf("Registered ServiceWeaver services from: %T\n", s)
		}
	}

	return nil
}

// GetWeaverService 获取 ServiceWeaver 服务的通用方法
func (f *APIFramework) GetWeaverService(name string) (interface{}, error) {
	service, ok := f.weaverServices[name]
	if !ok {
		return nil, fmt.Errorf("service %s not found", name)
	}
	return service, nil
}

// MustGetWeaverService 获取 ServiceWeaver 服务并进行类型断言
func (f *APIFramework) MustGetWeaverService(name string, target interface{}) error {
	service, err := f.GetWeaverService(name)
	if err != nil {
		return err
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	serviceValue := reflect.ValueOf(service)
	if !serviceValue.Type().AssignableTo(targetValue.Elem().Type()) {
		return fmt.Errorf("service %s is not assignable to target type", name)
	}

	targetValue.Elem().Set(serviceValue)
	return nil
}
