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

func mustResolve[T any](tb testing.TB) *jsonschema.Resolved {
	tb.Helper()
	s, err := jsonschema.For[T](nil)
	if err != nil {
		tb.Fatalf("jsonschema.For[%T]: %v", *new(T), err)
	}
	r, err := s.Resolve(nil)
	if err != nil {
		tb.Fatalf("Resolve: %v", err)
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

// TestIsJSONSafe checks the correctness of the isJSONSafe helper function.
func TestIsJSONSafe(t *testing.T) {
	tests := []struct {
		val  any
		want bool
	}{
		{float64(42.5), true},
		{"test", true},
		{true, true},
		{nil, true},
		{123, false}, // Go int is not safe because JSON float64 is the canonical number type
		{float32(42.5), false},
		{[]string{"a"}, false},
		{map[string]any{"a": 1}, false},
	}

	for _, tc := range tests {
		if got := isJSONSafe(tc.val); got != tc.want {
			t.Errorf("isJSONSafe(%v (%T)) = %v, want %v", tc.val, tc.val, got, tc.want)
		}
	}
}

// BenchmarkConvertTo_String_Optimized benchmarks the fast-path ConvertToWithJSONSchema
// using a safe scalar value.
func BenchmarkConvertTo_String_Optimized(b *testing.B) {
	schema := mustResolve[string](b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[string, string]("hello", schema)
	}
}

// BenchmarkConvertTo_Struct_SlowPath benchmarks the conversion of a complex struct
// which cannot use the fast-path.
func BenchmarkConvertTo_Struct_SlowPath(b *testing.B) {
	type complexStruct struct {
		Field string `json:"field"`
	}
	schema := mustResolve[complexStruct](b)
	val := complexStruct{Field: "hello"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[complexStruct, complexStruct](val, schema)
	}
}

// BenchmarkValidate_String_Optimized benchmarks the fast-path ValidateWithJSONSchema.
func BenchmarkValidate_String_Optimized(b *testing.B) {
	schema := mustResolve[string](b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateWithJSONSchema("hello", schema)
	}
}

// BenchmarkValidate_Struct_SlowPath benchmarks the validation of a complex struct
// which cannot use the fast-path.
func BenchmarkValidate_Struct_SlowPath(b *testing.B) {
	type complexStruct struct {
		Field string `json:"field"`
	}
	schema := mustResolve[complexStruct](b)
	val := complexStruct{Field: "hello"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateWithJSONSchema(val, schema)
	}
}
