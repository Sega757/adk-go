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
	t.Run("string bypass", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[string, string]("hello", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("bool bypass", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[bool, bool](true, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("got %t, want true", got)
		}
	})

	t.Run("float64 bypass", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[float64, float64](45.6, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 45.6 {
			t.Errorf("got %f, want 45.6", got)
		}
	})

	t.Run("nil bypass", func(t *testing.T) {
		var val *int
		got, err := ConvertToWithJSONSchema[*int, *int](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("fallback normal map", func(t *testing.T) {
		val := map[string]any{"key": "value"}
		got, err := ConvertToWithJSONSchema[map[string]any, map[string]any](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, val) {
			t.Errorf("got %v, want %v", got, val)
		}
	})

	t.Run("untyped nil bypass", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[any, any](nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_SafeType(b *testing.B) {
	val := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertToWithJSONSchema[string, string](val, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertToWithJSONSchema_UnsafeType(b *testing.B) {
	val := map[string]any{"key": "value"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertToWithJSONSchema[map[string]any, map[string]any](val, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
