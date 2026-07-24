// Copyright 2026 Google LLC
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

	"github.com/google/jsonschema-go/jsonschema"
)

func mustResolve[T any](t *testing.T) *jsonschema.Resolved {
	if t != nil {
		t.Helper()
	}
	s, err := jsonschema.For[T](nil)
	if err != nil {
		if t != nil {
			t.Fatalf("jsonschema.For[%T]: %v", *new(T), err)
		} else {
			panic(err)
		}
	}
	r, err := s.Resolve(nil)
	if err != nil {
		if t != nil {
			t.Fatalf("Resolve: %v", err)
		} else {
			panic(err)
		}
	}
	return r
}

// TestConvertToWithJSONSchema_NilInputObjectSchema checks that a nil
// input (a tool invoked with no arguments) validates against an object
// schema.
func TestConvertToWithJSONSchema_NilInputObjectSchema(t *testing.T) {
	schema := mustResolve[map[string]any](t)

	var in map[string]any
	got, err := ConvertToWithJSONSchema[map[string]any, map[string]any](in, schema)
	if err != nil {
		t.Fatalf("ConvertToWithJSONSchema(nil map, object schema) returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want nil/empty map", got)
	}
}

// TestConvertToWithJSONSchema_ScalarInputs checks that scalar and
// array inputs validate against matching non-object schemas.
func TestConvertToWithJSONSchema_ScalarInputs(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		schema := mustResolve[string](t)
		got, err := ConvertToWithJSONSchema[string, string]("hello", schema)
		if err != nil {
			t.Fatalf("string input against string schema: %v", err)
		}
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("slice", func(t *testing.T) {
		schema := mustResolve[[]int](t)
		got, err := ConvertToWithJSONSchema[[]int, []int]([]int{1, 2, 3}, schema)
		if err != nil {
			t.Fatalf("slice input against array schema: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("got %v, want [1 2 3]", got)
		}
	})
}

// TestConvertToWithJSONSchema_NilInputNonObjectSchema checks that a nil
// input is rejected by a non-object schema rather than coerced.
func TestConvertToWithJSONSchema_NilInputNonObjectSchema(t *testing.T) {
	schema := mustResolve[string](t)

	var in *string
	_, err := ConvertToWithJSONSchema[*string, *string](in, schema)
	if err == nil {
		t.Fatal("expected validation error for null against string schema, got nil")
	}
}

// TestConvertToWithJSONSchema_TypeMismatchStillFails checks that a
// non-null value of the wrong type is still rejected by an object
// schema.
func TestConvertToWithJSONSchema_TypeMismatchStillFails(t *testing.T) {
	schema := mustResolve[map[string]any](t)

	_, err := ConvertToWithJSONSchema[string, map[string]any]("not-an-object", schema)
	if err == nil {
		t.Fatal("expected validation error for string against object schema, got nil")
	}
}

// TestConvertToWithJSONSchema_NoSchemaSkipsValidation verifies that a
// nil schema bypasses validation entirely.
func TestConvertToWithJSONSchema_NoSchemaSkipsValidation(t *testing.T) {
	got, err := ConvertToWithJSONSchema[map[string]any, map[string]any](nil, nil)
	if err != nil {
		t.Fatalf("nil schema should skip validation, got error: %v", err)
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

// TestConvertToWithJSONSchema_JSONSafeBypass verifies that JSON-safe types and untyped nil
// bypass standard marshal/unmarshal but still validate properly, while other types safely
// fall back to standard JSON marshal/unmarshal.
func TestConvertToWithJSONSchema_JSONSafeBypass(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		schema := mustResolve[float64](t)
		got, err := ConvertToWithJSONSchema[float64, float64](123.45, schema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 123.45 {
			t.Errorf("got %v, want 123.45", got)
		}
	})

	t.Run("string", func(t *testing.T) {
		schema := mustResolve[string](t)
		got, err := ConvertToWithJSONSchema[string, string]("test-str", schema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "test-str" {
			t.Errorf("got %v, want test-str", got)
		}
	})

	t.Run("bool", func(t *testing.T) {
		schema := mustResolve[bool](t)
		got, err := ConvertToWithJSONSchema[bool, bool](true, schema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		got, err := ConvertToWithJSONSchema[*string, *string](nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("typed nil pointer fallback", func(t *testing.T) {
		schema := mustResolve[*string](t)
		var p *string
		got, err := ConvertToWithJSONSchema[*string, *string](p, schema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("fallback int", func(t *testing.T) {
		schema := mustResolve[int](t)
		got, err := ConvertToWithJSONSchema[int, int](42, schema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42 {
			t.Errorf("got %v, want 42", got)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_Float64_Optimized(b *testing.B) {
	schema := mustResolve[float64](nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[float64, float64](12.34, schema)
	}
}

func BenchmarkConvertToWithJSONSchema_Int_Fallback(b *testing.B) {
	schema := mustResolve[int](nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[int, int](42, schema)
	}
}

func BenchmarkValidateWithJSONSchema_String_Optimized(b *testing.B) {
	schema := mustResolve[string](nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateWithJSONSchema("hello", schema)
	}
}

func BenchmarkValidateWithJSONSchema_Struct_Fallback(b *testing.B) {
	type simple struct {
		X int `json:"x"`
	}
	schema := mustResolve[simple](nil)
	val := simple{X: 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateWithJSONSchema(val, schema)
	}
}
