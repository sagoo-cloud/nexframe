package gvar

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestVar(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		safe     bool
		expected interface{}
		testFunc func(*testing.T, *Var)
	}{
		{
			name:     "String",
			value:    "test",
			expected: "test",
			testFunc: func(t *testing.T, v *Var) {
				if v.String() != "test" {
					t.Errorf("Expected 'test', got %v", v.String())
				}
			},
		},
		{
			name:     "Int",
			value:    42,
			expected: 42,
			testFunc: func(t *testing.T, v *Var) {
				if v.Int() != 42 {
					t.Errorf("Expected 42, got %v", v.Int())
				}
			},
		},
		{
			name:     "Bool",
			value:    true,
			expected: true,
			testFunc: func(t *testing.T, v *Var) {
				if !v.Bool() {
					t.Errorf("Expected true, got %v", v.Bool())
				}
			},
		},
		{
			name:     "Float64",
			value:    3.14,
			expected: 3.14,
			testFunc: func(t *testing.T, v *Var) {
				if v.Float64() != 3.14 {
					t.Errorf("Expected 3.14, got %v", v.Float64())
				}
			},
		},
		{
			name:  "Time",
			value: "2024-03-01 12:00:00",
			testFunc: func(t *testing.T, v *Var) {
				expected := time.Date(2024, 3, 1, 12, 0, 0, 0, time.Local)
				if !v.Time().Equal(expected) {
					t.Errorf("Expected %v, got %v", expected, v.Time())
				}
			},
		},
		{
			name:  "Bytes",
			value: []byte("hello"),
			testFunc: func(t *testing.T, v *Var) {
				if string(v.Bytes()) != "hello" {
					t.Errorf("Expected 'hello', got %v", string(v.Bytes()))
				}
			},
		},
		{
			name:  "Interface",
			value: map[string]int{"a": 1},
			testFunc: func(t *testing.T, v *Var) {
				expected := map[string]int{"a": 1}
				if !reflect.DeepEqual(v.Interface(), expected) {
					t.Errorf("Expected %v, got %v", expected, v.Interface())
				}
			},
		},
		{
			name:  "MarshalJSON",
			value: map[string]string{"key": "value"},
			testFunc: func(t *testing.T, v *Var) {
				bytes, err := json.Marshal(v)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				expected := `{"key":"value"}`
				if string(bytes) != expected {
					t.Errorf("Expected %s, got %s", expected, string(bytes))
				}
			},
		},
		{
			name:  "UnmarshalJSON",
			value: nil,
			testFunc: func(t *testing.T, v *Var) {
				jsonData := []byte(`{"key": "value"}`)
				err := json.Unmarshal(jsonData, v)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				expected := map[string]interface{}{"key": "value"}
				if !reflect.DeepEqual(v.Val(), expected) {
					t.Errorf("Expected %v, got %v", expected, v.Val())
				}
			},
		},
		{
			name:  "DeepCopy",
			value: []int{1, 2, 3},
			testFunc: func(t *testing.T, v *Var) {
				copied := v.DeepCopy().(*Var)
				if !reflect.DeepEqual(v.Val(), copied.Val()) {
					t.Errorf("DeepCopy failed, expected %v, got %v", v.Val(), copied.Val())
				}
				// Modify the original to ensure deep copy
				v.Set([]int{4, 5, 6})
				if reflect.DeepEqual(v.Val(), copied.Val()) {
					t.Errorf("DeepCopy failed, values should be different after modification")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(tt.value, tt.safe)
			defer v.Release()

			if tt.testFunc != nil {
				tt.testFunc(t, v)
			}
		})
	}
}

func TestVarClone(t *testing.T) {
	original := New(42)
	clone := original.Clone()

	if original.Int() != clone.Int() {
		t.Errorf("Clone failed, expected %d, got %d", original.Int(), clone.Int())
	}

	original.Set(100)
	if clone.Int() == original.Int() {
		t.Errorf("Clone is not independent, both values are %d", original.Int())
	}
}

func TestVarConversions(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		testFunc func(*testing.T, *Var)
	}{
		{
			name:  "Int8",
			value: int8(8),
			testFunc: func(t *testing.T, v *Var) {
				if v.Int8() != 8 {
					t.Errorf("Expected 8, got %v", v.Int8())
				}
			},
		},
		{
			name:  "Int16",
			value: int16(16),
			testFunc: func(t *testing.T, v *Var) {
				if v.Int16() != 16 {
					t.Errorf("Expected 16, got %v", v.Int16())
				}
			},
		},
		{
			name:  "Int32",
			value: int32(32),
			testFunc: func(t *testing.T, v *Var) {
				if v.Int32() != 32 {
					t.Errorf("Expected 32, got %v", v.Int32())
				}
			},
		},
		{
			name:  "Int64",
			value: int64(64),
			testFunc: func(t *testing.T, v *Var) {
				if v.Int64() != 64 {
					t.Errorf("Expected 64, got %v", v.Int64())
				}
			},
		},
		{
			name:  "Uint",
			value: uint(42),
			testFunc: func(t *testing.T, v *Var) {
				if v.Uint() != 42 {
					t.Errorf("Expected 42, got %v", v.Uint())
				}
			},
		},
		{
			name:  "Uint8",
			value: uint8(8),
			testFunc: func(t *testing.T, v *Var) {
				if v.Uint8() != 8 {
					t.Errorf("Expected 8, got %v", v.Uint8())
				}
			},
		},
		{
			name:  "Uint16",
			value: uint16(16),
			testFunc: func(t *testing.T, v *Var) {
				if v.Uint16() != 16 {
					t.Errorf("Expected 16, got %v", v.Uint16())
				}
			},
		},
		{
			name:  "Uint32",
			value: uint32(32),
			testFunc: func(t *testing.T, v *Var) {
				if v.Uint32() != 32 {
					t.Errorf("Expected 32, got %v", v.Uint32())
				}
			},
		},
		{
			name:  "Uint64",
			value: uint64(64),
			testFunc: func(t *testing.T, v *Var) {
				if v.Uint64() != 64 {
					t.Errorf("Expected 64, got %v", v.Uint64())
				}
			},
		},
		{
			name:  "Float32",
			value: float32(3.14),
			testFunc: func(t *testing.T, v *Var) {
				if v.Float32() != float32(3.14) {
					t.Errorf("Expected 3.14, got %v", v.Float32())
				}
			},
		},
		{
			name:  "Duration",
			value: "1h30m",
			testFunc: func(t *testing.T, v *Var) {
				expected := 90 * time.Minute
				if v.Duration() != expected {
					t.Errorf("Expected %v, got %v", expected, v.Duration())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(tt.value)
			defer v.Release()

			if tt.testFunc != nil {
				tt.testFunc(t, v)
			}
		})
	}
}
