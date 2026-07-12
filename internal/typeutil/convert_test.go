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
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/internal/typeutil"
)

type BenchmarkArgs struct {
	A float64 `json:"a"`
	B string  `json:"b"`
}

func BenchmarkConvertToWithJSONSchema_Map(b *testing.B) {
	schema, err := jsonschema.For[BenchmarkArgs](nil)
	if err != nil {
		b.Fatalf("failed to create schema: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		b.Fatalf("failed to resolve schema: %v", err)
	}

	input := map[string]any{
		"a": float64(42),
		"b": "hello",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := typeutil.ConvertToWithJSONSchema[map[string]any, BenchmarkArgs](input, resolved)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkConvertToWithJSONSchema_Struct(b *testing.B) {
	schema, err := jsonschema.For[BenchmarkArgs](nil)
	if err != nil {
		b.Fatalf("failed to create schema: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		b.Fatalf("failed to resolve schema: %v", err)
	}

	input := BenchmarkArgs{
		A: 42,
		B: "hello",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := typeutil.ConvertToWithJSONSchema[BenchmarkArgs, map[string]any](input, resolved)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestConvertToWithJSONSchema_Safety(t *testing.T) {
	type Custom struct {
		Date time.Time `json:"date"`
	}

	schema, err := jsonschema.For[Custom](nil)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	// input map with non-JSON primitive type (time.Time)
	now := time.Now()
	input := map[string]any{
		"date": now,
	}

	// This should run successfully because our optimizer safely detects
	// that time.Time is not a standard JSON type and falls back to
	// standard marshal/unmarshal before validating.
	got, err := typeutil.ConvertToWithJSONSchema[map[string]any, Custom](input, resolved)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if !got.Date.Equal(now.Truncate(time.Nanosecond)) {
		t.Errorf("expected date %v, got %v", now, got.Date)
	}
}

func TestConvertToWithJSONSchema_IntFallsBackToMarshal(t *testing.T) {
	schema, err := jsonschema.For[BenchmarkArgs](nil)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	// Map with an integer. Since integer is not "JSON safe" in isJSONSafe,
	// it should fall back to marshal/unmarshal cycle and validate/convert correctly.
	input := map[string]any{
		"a": 42, // int type
		"b": "hello",
	}

	got, err := typeutil.ConvertToWithJSONSchema[map[string]any, BenchmarkArgs](input, resolved)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if got.A != 42 {
		t.Errorf("expected 42, got %v", got.A)
	}
}

func TestConvertToWithJSONSchema_NoSharedReferences(t *testing.T) {
	schema, err := jsonschema.For[BenchmarkArgs](nil)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	input := map[string]any{
		"a": float64(42),
		"b": "hello",
	}

	got, err := typeutil.ConvertToWithJSONSchema[map[string]any, map[string]any](input, resolved)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Verify that mutating the returned map does not affect the input map.
	got["b"] = "mutated"
	if input["b"] == "mutated" {
		t.Error("returned map shares references with input map, leading to mutation/data-race risk")
	}
}
