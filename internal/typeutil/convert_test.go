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

package typeutil_test

import (
	"testing"

	"google.golang.org/adk/internal/typeutil"
)

func TestConvertToWithJSONSchema(t *testing.T) {
	t.Run("primitive string bypass", func(t *testing.T) {
		val, err := typeutil.ConvertToWithJSONSchema[string, string]("hello", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "hello" {
			t.Errorf("expected %q, got %q", "hello", val)
		}
	})

	t.Run("primitive float64 bypass", func(t *testing.T) {
		val, err := typeutil.ConvertToWithJSONSchema[float64, float64](12.34, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 12.34 {
			t.Errorf("expected %v, got %v", 12.34, val)
		}
	})

	t.Run("primitive bool bypass", func(t *testing.T) {
		val, err := typeutil.ConvertToWithJSONSchema[bool, bool](true, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != true {
			t.Errorf("expected %v, got %v", true, val)
		}
	})

	t.Run("untyped nil bypass", func(t *testing.T) {
		val, err := typeutil.ConvertToWithJSONSchema[any, any](nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil, got %v", val)
		}
	})

	t.Run("typed nil fallback", func(t *testing.T) {
		var p *int
		val, err := typeutil.ConvertToWithJSONSchema[*int, *int](p, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil, got %v", val)
		}
	})

	t.Run("struct conversion standard path", func(t *testing.T) {
		type S struct {
			A int `json:"a"`
		}
		v := S{A: 42}
		val, err := typeutil.ConvertToWithJSONSchema[S, map[string]any](v, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val["a"] != float64(42) {
			t.Errorf("expected 42, got %v", val["a"])
		}
	})
}

func BenchmarkConvertToWithJSONSchema_Bypass(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typeutil.ConvertToWithJSONSchema[string, string]("hello", nil)
	}
}

func BenchmarkConvertToWithJSONSchema_Fallback(b *testing.B) {
	type S struct {
		A int `json:"a"`
	}
	v := S{A: 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typeutil.ConvertToWithJSONSchema[S, map[string]any](v, nil)
	}
}
