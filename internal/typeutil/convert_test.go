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
	"reflect"
	"testing"
)

func TestConvertToWithJSONSchema(t *testing.T) {
	t.Run("safe type string", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[string, string]("hello", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("safe type float64", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[float64, float64](3.14, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 3.14 {
			t.Errorf("got %f, want %f", got, 3.14)
		}
	})

	t.Run("safe type bool", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[bool, bool](true, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != true {
			t.Errorf("got %t, want true", got)
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[any, any](nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("unsafe type int", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[int, int](123, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 123 {
			t.Errorf("got %d, want 123", got)
		}
	})

	t.Run("unsafe type struct", func(t *testing.T) {
		type myStruct struct {
			Field string `json:"field"`
		}
		v := myStruct{Field: "value"}
		got, err := ConvertToWithJSONSchema[myStruct, myStruct](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, v) {
			t.Errorf("got %+v, want %+v", got, v)
		}
	})

	t.Run("typed nil pointer fallback", func(t *testing.T) {
		var p *int
		got, err := ConvertToWithJSONSchema[*int, *int](p, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_SafeType(b *testing.B) {
	b.Run("string bypass", func(b *testing.B) {
		v := "benchmark_test_string"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertToWithJSONSchema[string, string](v, nil)
		}
	})

	b.Run("float64 bypass", func(b *testing.B) {
		v := 1234.56
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertToWithJSONSchema[float64, float64](v, nil)
		}
	})

	b.Run("bool bypass", func(b *testing.B) {
		v := true
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertToWithJSONSchema[bool, bool](v, nil)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_Fallback(b *testing.B) {
	b.Run("int fallback", func(b *testing.B) {
		v := 123
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertToWithJSONSchema[int, int](v, nil)
		}
	})

	b.Run("struct fallback", func(b *testing.B) {
		type myStruct struct {
			Field string `json:"field"`
		}
		v := myStruct{Field: "value"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertToWithJSONSchema[myStruct, myStruct](v, nil)
		}
	})
}
