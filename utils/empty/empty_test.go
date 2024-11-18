package empty

import (
	"testing"
	"time"
)

// 测试基础类型
func TestIsEmptyBasicTypes(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil值", nil, true},
		{"空字符串", "", true},
		{"非空字符串", "hello", false},
		{"零整数", 0, true},
		{"非零整数", 42, false},
		{"零浮点数", 0.0, true},
		{"非零浮点数", 3.14, false},
		{"false布尔值", false, true},
		{"true布尔值", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.value); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试复合类型
func TestIsEmptyCompositeTypes(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"空切片", []int{}, true},
		{"非空切片", []int{1, 2, 3}, false},
		{"空map", map[string]int{}, true},
		{"非空map", map[string]int{"a": 1}, false},
		{"空数组", [0]int{}, true},
		{"非空数组", [3]int{1, 2, 3}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.value); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试指针类型
func TestIsEmptyPointers(t *testing.T) {
	var nilPtr *int
	value := 42
	ptr := &value

	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil指针", nilPtr, true},
		{"非nil指针", ptr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.value); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试结构体
type testStruct struct {
	Name string
	Age  int
}

func TestIsEmptyStructs(t *testing.T) {
	emptyTime := time.Time{}
	nonEmptyTime := time.Now()

	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"空结构体", testStruct{}, true},
		{"非空结构体", testStruct{Name: "test", Age: 20}, false},
		{"空time.Time", emptyTime, true},
		{"非空time.Time", nonEmptyTime, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.value); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试接口
type testInterface interface {
	Test() string
}

type testImplementation struct{}

func (t testImplementation) Test() string { return "test" }

func TestIsEmptyInterfaces(t *testing.T) {
	var nilIface testInterface
	nonNilIface := testImplementation{}

	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil接口", nilIface, true},
		{"非nil接口", nonNilIface, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.value); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试IsNil函数
func TestIsNil(t *testing.T) {
	var nilPtr *int
	value := 42
	ptr := &value

	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil值", nil, true},
		{"nil指针", nilPtr, true},
		{"非nil指针", ptr, false},
		{"非指针类型", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.value); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试IsZero函数
func TestIsZero(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"零整数", 0, true},
		{"非零整数", 42, false},
		{"零字符串", "", true},
		{"非零字符串", "hello", false},
		{"空结构体", testStruct{}, true},
		{"非空结构体", testStruct{Name: "test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZero(tt.value); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试IsEmptyOrZero函数
func TestIsEmptyOrZero(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil值", nil, true},
		{"零值", 0, true},
		{"空字符串", "", true},
		{"非空非零值", 42, false},
		{"非空字符串", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmptyOrZero(tt.value); got != tt.want {
				t.Errorf("IsEmptyOrZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试循环引用情况
func TestCircularReferences(t *testing.T) {
	type circular struct {
		Self *circular
	}

	c := &circular{}
	c.Self = c

	if IsEmpty(c) {
		t.Error("Circular reference should not be considered empty")
	}
}

// 测试并发安全性
func TestConcurrentAccess(t *testing.T) {
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			IsEmpty(map[string]int{"test": 1})
			IsNil(nil)
			IsZero(0)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
