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

func TestConvertToWithJSONSchema_Bypass(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		v := "hello"
		got, err := ConvertToWithJSONSchema[string, string](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != v {
			t.Errorf("got %q, want %q", got, v)
		}
	})

	t.Run("float64", func(t *testing.T) {
		v := 123.45
		got, err := ConvertToWithJSONSchema[float64, float64](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != v {
			t.Errorf("got %v, want %v", got, v)
		}
	})

	t.Run("bool", func(t *testing.T) {
		v := true
		got, err := ConvertToWithJSONSchema[bool, bool](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != v {
			t.Errorf("got %v, want %v", got, v)
		}
	})

	t.Run("nil interface", func(t *testing.T) {
		var v any = nil
		got, err := ConvertToWithJSONSchema[any, any](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("typed nil pointer (should fallback to marshal)", func(t *testing.T) {
		// A typed nil pointer converted to `any` should result in untyped `nil`.
		var v *string
		got, err := ConvertToWithJSONSchema[*string, any](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("int fallthrough", func(t *testing.T) {
		v := 123
		got, err := ConvertToWithJSONSchema[int, int](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != v {
			t.Errorf("got %v, want %v", got, v)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_WithBypass(b *testing.B) {
	v := "hello"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[string, string](v, nil)
	}
}

func BenchmarkConvertToWithJSONSchema_WithoutBypass(b *testing.B) {
	v := []string{"hello", "world"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[[]string, []string](v, nil)
	}
}
