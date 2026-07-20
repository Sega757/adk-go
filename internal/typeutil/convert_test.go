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

func TestConvertToWithJSONSchema(t *testing.T) {
	t.Run("safe float64", func(t *testing.T) {
		val := 12.34
		got, err := ConvertToWithJSONSchema[float64, float64](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != val {
			t.Errorf("got %v, want %v", got, val)
		}
	})

	t.Run("safe string", func(t *testing.T) {
		val := "hello"
		got, err := ConvertToWithJSONSchema[string, string](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != val {
			t.Errorf("got %v, want %v", got, val)
		}
	})

	t.Run("safe bool", func(t *testing.T) {
		val := true
		got, err := ConvertToWithJSONSchema[bool, bool](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != val {
			t.Errorf("got %v, want %v", got, val)
		}
	})

	t.Run("unsafe struct slow path", func(t *testing.T) {
		type Foo struct {
			Bar string `json:"bar"`
		}
		type Baz struct {
			Bar string `json:"bar"`
		}
		val := Foo{Bar: "test"}
		got, err := ConvertToWithJSONSchema[Foo, Baz](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Bar != val.Bar {
			t.Errorf("got %v, want %v", got.Bar, val.Bar)
		}
	})

	t.Run("unsafe type assertion mismatch slow path", func(t *testing.T) {
		// Go int is not floats in JSON, needs standard flow to handle conversion via JSON
		val := int(42)
		got, err := ConvertToWithJSONSchema[int, float64](val, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42.0 {
			t.Errorf("got %v, want %v", got, 42.0)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_FastPath(b *testing.B) {
	val := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[string, string](val, nil)
	}
}

func BenchmarkConvertToWithJSONSchema_SlowPath(b *testing.B) {
	type Foo struct {
		Bar string `json:"bar"`
	}
	type Baz struct {
		Bar string `json:"bar"`
	}
	val := Foo{Bar: "hello world"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[Foo, Baz](val, nil)
	}
}
