// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package typeutil

import (
	"testing"
)

func TestConvertToWithJSONSchema_SafeTypes(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		val := float64(42.5)
		got, err := ConvertToWithJSONSchema[float64, any](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != val {
			t.Errorf("got %v, want %v", got, val)
		}
	})

	t.Run("string", func(t *testing.T) {
		val := "hello"
		got, err := ConvertToWithJSONSchema[string, string](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != val {
			t.Errorf("got %v, want %v", got, val)
		}
	})

	t.Run("bool", func(t *testing.T) {
		val := true
		got, err := ConvertToWithJSONSchema[bool, any](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != val {
			t.Errorf("got %v, want %v", got, val)
		}
	})

	t.Run("nil", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[any, any](nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}

func TestConvertToWithJSONSchema_Fallback(t *testing.T) {
	t.Run("int to float64 via any", func(t *testing.T) {
		// When converting an int through any, JSON marshal/unmarshal
		// converts it to float64, whereas type assertion would keep it as int.
		// Fallback should be triggered so it gets correctly unmarshaled as float64.
		val := int(42)
		got, err := ConvertToWithJSONSchema[int, any](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f, ok := got.(float64); !ok || f != 42.0 {
			t.Errorf("got %T(%v), want float64(42.0)", got, got)
		}
	})

	t.Run("struct to map", func(t *testing.T) {
		type Simple struct {
			Name string `json:"name"`
		}
		val := Simple{Name: "test"}
		got, err := ConvertToWithJSONSchema[Simple, map[string]any](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["name"] != "test" {
			t.Errorf("got %v, want 'test'", got["name"])
		}
	})
}

func BenchmarkConvertToWithJSONSchema_String(b *testing.B) {
	val := "benchmark_test_string"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[string, string](val, nil)
	}
}

func BenchmarkConvertToWithJSONSchema_Float64(b *testing.B) {
	val := float64(123.45)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[float64, float64](val, nil)
	}
}

func BenchmarkConvertToWithJSONSchema_Struct(b *testing.B) {
	type Simple struct {
		Name string `json:"name"`
	}
	val := Simple{Name: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[Simple, map[string]any](val, nil)
	}
}
