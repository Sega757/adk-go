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

// TestIsJSONSafe verifies the behavior of the isJSONSafe utility.
func TestIsJSONSafe(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"float64", float64(3.14), true},
		{"string", "test", true},
		{"bool", true, true},
		{"untyped nil", nil, true},
		{"int (not safe)", int(42), false},
		{"typed nil pointer (not safe)", (*string)(nil), false},
		{"map (not safe)", map[string]any{}, false},
		{"slice (not safe)", []any{}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isJSONSafe(tc.input)
			if got != tc.expected {
				t.Errorf("isJSONSafe(%T(%v)) = %v; want %v", tc.input, tc.input, got, tc.expected)
			}
		})
	}
}

// BenchmarkConvertToWithJSONSchema_Bypass measures performance of ConvertToWithJSONSchema
// when the input is JSON-safe and conversion type matches exactly.
func BenchmarkConvertToWithJSONSchema_Bypass(b *testing.B) {
	schema, err := jsonschema.For[string](nil)
	if err != nil {
		b.Fatalf("jsonschema.For: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		b.Fatalf("Resolve: %v", err)
	}

	input := "benchmark-test-string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertToWithJSONSchema[string, string](input, resolved)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConvertToWithJSONSchema_NoBypass measures performance of ConvertToWithJSONSchema
// when bypass cannot be used (e.g. converting a struct or map).
func BenchmarkConvertToWithJSONSchema_NoBypass(b *testing.B) {
	type TestStruct struct {
		Field string `json:"field"`
	}

	schema, err := jsonschema.For[TestStruct](nil)
	if err != nil {
		b.Fatalf("jsonschema.For: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		b.Fatalf("Resolve: %v", err)
	}

	input := TestStruct{Field: "benchmark-test-string"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertToWithJSONSchema[TestStruct, TestStruct](input, resolved)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateWithJSONSchema_Bypass measures performance of ValidateWithJSONSchema
// when the input is JSON-safe.
func BenchmarkValidateWithJSONSchema_Bypass(b *testing.B) {
	schema, err := jsonschema.For[string](nil)
	if err != nil {
		b.Fatalf("jsonschema.For: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		b.Fatalf("Resolve: %v", err)
	}

	input := "benchmark-test-string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateWithJSONSchema(input, resolved)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateWithJSONSchema_NoBypass measures performance of ValidateWithJSONSchema
// when bypass cannot be used.
func BenchmarkValidateWithJSONSchema_NoBypass(b *testing.B) {
	type TestStruct struct {
		Field string `json:"field"`
	}

	schema, err := jsonschema.For[TestStruct](nil)
	if err != nil {
		b.Fatalf("jsonschema.For: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		b.Fatalf("Resolve: %v", err)
	}

	input := TestStruct{Field: "benchmark-test-string"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateWithJSONSchema(input, resolved)
		if err != nil {
			b.Fatal(err)
		}
	}
}
