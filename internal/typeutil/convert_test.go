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
		v := "hello"
		got, err := ConvertToWithJSONSchema[string, string](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("safe type float64", func(t *testing.T) {
		v := 123.45
		got, err := ConvertToWithJSONSchema[float64, float64](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 123.45 {
			t.Errorf("got %f, want %f", got, 123.45)
		}
	})

	t.Run("safe type bool", func(t *testing.T) {
		v := true
		got, err := ConvertToWithJSONSchema[bool, bool](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("got %t, want %t", got, true)
		}
	})

	t.Run("safe type nil", func(t *testing.T) {
		var v *int
		got, err := ConvertToWithJSONSchema[*int, *int](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("int to float64 check", func(t *testing.T) {
		// Int should NOT be bypassed because we want it to be converted to float64
		// if the target is any (which mimics JSON behavior).
		v := 123
		got, err := ConvertToWithJSONSchema[int, any](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reflect.TypeOf(got).Kind() != reflect.Float64 {
			t.Errorf("got type %T, want float64", got)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_SafeString(b *testing.B) {
	v := "hello"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[string, string](v, nil)
	}
}

func BenchmarkConvertToWithJSONSchema_SafeFloat64(b *testing.B) {
	v := 123.45
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[float64, float64](v, nil)
	}
}

func BenchmarkConvertToWithJSONSchema_UnsafeInt(b *testing.B) {
	v := 123
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[int, any](v, nil)
	}
}
