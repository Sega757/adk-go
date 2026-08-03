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
	t.Helper()
	s, err := jsonschema.For[T](nil)
	if err != nil {
		t.Fatalf("jsonschema.For[%T]: %v", *new(T), err)
	}
	r, err := s.Resolve(nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
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

// TestConvertToWithJSONSchema_FastPaths checks the fast-path for string, float64, bool, and untyped nil.
func TestConvertToWithJSONSchema_FastPaths(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		schema := mustResolve[string](t)
		got, err := ConvertToWithJSONSchema[string, string]("test", schema)
		if err != nil {
			t.Fatalf("fast-path string conversion failed: %v", err)
		}
		if got != "test" {
			t.Errorf("got %q, want 'test'", got)
		}
	})

	t.Run("float64", func(t *testing.T) {
		schema := mustResolve[float64](t)
		got, err := ConvertToWithJSONSchema[float64, float64](123.45, schema)
		if err != nil {
			t.Fatalf("fast-path float64 conversion failed: %v", err)
		}
		if got != 123.45 {
			t.Errorf("got %f, want 123.45", got)
		}
	})

	t.Run("bool", func(t *testing.T) {
		schema := mustResolve[bool](t)
		got, err := ConvertToWithJSONSchema[bool, bool](true, schema)
		if err != nil {
			t.Fatalf("fast-path bool conversion failed: %v", err)
		}
		if !got {
			t.Errorf("got %t, want true", got)
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		schema := mustResolve[any](t)
		got, err := ConvertToWithJSONSchema[any, any](nil, schema)
		if err != nil {
			t.Fatalf("fast-path untyped nil conversion failed: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("typed nil pointer", func(t *testing.T) {
		schema := mustResolve[*string](t)
		var in *string
		got, err := ConvertToWithJSONSchema[*string, *string](in, schema)
		if err != nil {
			t.Fatalf("typed nil pointer conversion failed: %v", err)
		}
		if got != nil {
			t.Errorf("got %v, want nil pointer", got)
		}
	})
}

// TestValidateWithJSONSchema_FastPaths checks ValidateWithJSONSchema fast-paths.
func TestValidateWithJSONSchema_FastPaths(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		schema := mustResolve[string](t)
		if err := ValidateWithJSONSchema("test", schema); err != nil {
			t.Fatalf("fast-path string validation failed: %v", err)
		}
	})

	t.Run("float64", func(t *testing.T) {
		schema := mustResolve[float64](t)
		if err := ValidateWithJSONSchema(123.45, schema); err != nil {
			t.Fatalf("fast-path float64 validation failed: %v", err)
		}
	})

	t.Run("bool", func(t *testing.T) {
		schema := mustResolve[bool](t)
		if err := ValidateWithJSONSchema(true, schema); err != nil {
			t.Fatalf("fast-path bool validation failed: %v", err)
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		schema := mustResolve[any](t)
		if err := ValidateWithJSONSchema(nil, schema); err != nil {
			t.Fatalf("fast-path untyped nil validation failed: %v", err)
		}
	})

	t.Run("typed nil pointer", func(t *testing.T) {
		schema := mustResolve[*string](t)
		var in *string
		if err := ValidateWithJSONSchema(in, schema); err != nil {
			t.Fatalf("typed nil pointer validation failed: %v", err)
		}
	})
}

func BenchmarkConvertToWithJSONSchema_FastPath_String(b *testing.B) {
	s, _ := jsonschema.For[string](nil)
	schema, _ := s.Resolve(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[string, string]("hello", schema)
	}
}

func BenchmarkConvertToWithJSONSchema_SlowPath_Struct(b *testing.B) {
	type testStruct struct {
		Field string `json:"field"`
	}
	s, _ := jsonschema.For[testStruct](nil)
	schema, _ := s.Resolve(nil)
	val := testStruct{Field: "hello"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[testStruct, testStruct](val, schema)
	}
}

func BenchmarkValidateWithJSONSchema_FastPath_String(b *testing.B) {
	s, _ := jsonschema.For[string](nil)
	schema, _ := s.Resolve(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateWithJSONSchema("hello", schema)
	}
}

func BenchmarkValidateWithJSONSchema_SlowPath_Struct(b *testing.B) {
	type testStruct struct {
		Field string `json:"field"`
	}
	s, _ := jsonschema.For[testStruct](nil)
	schema, _ := s.Resolve(nil)
	val := testStruct{Field: "hello"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateWithJSONSchema(val, schema)
	}
}
