package gstructs

import (
	"reflect"
	"sync"
	"testing"
)

// 测试用的结构体
type User struct {
	Id       int     `json:"id" valid:"required"`
	Name     string  `json:"name" valid:"required" description:"用户名称"`
	Age      int     `json:"age" valid:"required,min=0,max=150" description:"年龄"`
	Email    string  `json:"email" valid:"email" description:"邮箱"`
	Score    float64 `json:"score" default:"0.0" description:"得分"`
	IsActive bool    `json:"is_active" default:"true"`
	Tags     []string
	private  string // 未导出字段
}

// 嵌入结构体测试
type Employee struct {
	User           // 匿名嵌入
	Company string `json:"company" description:"公司名称"`
}

func TestField_Tag(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		tagKey    string
		want      string
	}{
		{
			name:      "获取json标签",
			fieldName: "Name",
			tagKey:    "json",
			want:      "name",
		},
		{
			name:      "获取description标签",
			fieldName: "Age",
			tagKey:    "description",
			want:      "年龄",
		},
		{
			name:      "获取默认值标签",
			fieldName: "Score",
			tagKey:    "default",
			want:      "0.0",
		},
		{
			name:      "获取不存在的标签",
			fieldName: "Name",
			tagKey:    "nonexistent",
			want:      "",
		},
	}

	typ := reflect.TypeOf(User{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := typ.FieldByName(tt.fieldName)
			if !ok {
				t.Fatalf("字段 %s 不存在", tt.fieldName)
			}

			f := &Field{
				Field: field,
				Value: reflect.New(field.Type).Elem(),
			}

			got := f.Tag(tt.tagKey)
			if got != tt.want {
				t.Errorf("Tag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFields(t *testing.T) {
	emp := &Employee{
		User: User{
			Id:       1,
			Name:     "John",
			Age:      30,
			Email:    "john@example.com",
			Score:    95.5,
			IsActive: true,
			Tags:     []string{"tag1", "tag2"},
		},
		Company: "Test Corp",
	}

	tests := []struct {
		name           string
		input          FieldsInput
		wantFieldCount int
		checkFields    []string
		shouldNotHave  []string
	}{
		{
			name: "非递归获取字段",
			input: FieldsInput{
				Pointer:         emp,
				RecursiveOption: RecursiveOptionNone,
			},
			wantFieldCount: 1, // 只有 Company，User 字段被当作匿名字段处理
			checkFields:    []string{"Company"},
			shouldNotHave:  []string{"private", "User"},
		},
		{
			name: "递归获取所有字段",
			input: FieldsInput{
				Pointer:         emp,
				RecursiveOption: RecursiveOptionEmbedded,
			},
			wantFieldCount: 8, // 所有导出字段
			checkFields:    []string{"Id", "Name", "Age", "Email", "Score", "IsActive", "Tags", "Company"},
			shouldNotHave:  []string{"private"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := Fields(tt.input)
			if err != nil {
				t.Fatalf("Fields() error = %v", err)
			}

			// 获取所有导出字段的名称
			exportedFields := make([]string, 0)
			for _, f := range fields {
				if f.IsExported() {
					exportedFields = append(exportedFields, f.Name())
				}
			}

			if len(exportedFields) != tt.wantFieldCount {
				t.Errorf("Fields() got %d fields, want %d. Exported fields: %v",
					len(exportedFields), tt.wantFieldCount, exportedFields)
			}

			// 检查必须存在的字段
			for _, expectedField := range tt.checkFields {
				found := false
				for _, field := range exportedFields {
					if field == expectedField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected field %q not found. Got fields: %v",
						expectedField, exportedFields)
				}
			}

			// 检查不应该存在的字段
			for _, notExpectedField := range tt.shouldNotHave {
				for _, field := range exportedFields {
					if field == notExpectedField {
						t.Errorf("Unexpected field %q found. Got fields: %v",
							notExpectedField, exportedFields)
					}
				}
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	routines := 10
	iterations := 10

	typ := reflect.TypeOf(User{})
	field, _ := typ.FieldByName("Name")
	f := &Field{
		Field: field,
		Value: reflect.New(field.Type).Elem(),
	}

	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = f.Tag("json")
				_ = f.Tag("description")
			}
		}()
	}

	wg.Wait()
}

func TestType_Methods(t *testing.T) {
	testType := Type{reflect.TypeOf(User{})}

	// 测试 Signature
	sig := testType.Signature()
	if sig == "" {
		t.Error("Signature() returned empty string")
	}

	// 测试 FieldKeys
	keys := testType.FieldKeys()
	expectedKeys := []string{"Id", "Name", "Age", "Email", "Score", "IsActive", "Tags", "private"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("FieldKeys() got %d keys, want %d", len(keys), len(expectedKeys))
	}

	// 验证字段名
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}
	for _, expectedKey := range expectedKeys {
		if !keyMap[expectedKey] {
			t.Errorf("Expected field %s not found in keys", expectedKey)
		}
	}
}

func BenchmarkField_Tag(b *testing.B) {
	typ := reflect.TypeOf(User{})
	field, _ := typ.FieldByName("Name")
	f := &Field{
		Field: field,
		Value: reflect.New(field.Type).Elem(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Tag("json")
	}
}

func BenchmarkField_TagMap(b *testing.B) {
	typ := reflect.TypeOf(User{})
	field, _ := typ.FieldByName("Age")
	f := &Field{
		Field: field,
		Value: reflect.New(field.Type).Elem(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.TagMap()
	}
}
