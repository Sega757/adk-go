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
	// Test string conversion (bypass)
	strVal, err := ConvertToWithJSONSchema[string, string]("hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strVal != "hello" {
		t.Errorf("expected 'hello', got %q", strVal)
	}

	// Test float64 conversion (bypass)
	floatVal, err := ConvertToWithJSONSchema[float64, float64](123.45, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if floatVal != 123.45 {
		t.Errorf("expected 123.45, got %v", floatVal)
	}

	// Test bool conversion (bypass)
	boolVal, err := ConvertToWithJSONSchema[bool, bool](true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !boolVal {
		t.Errorf("expected true, got %v", boolVal)
	}

	// Test nil conversion (bypass)
	anyVal, err := ConvertToWithJSONSchema[any, any](nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if anyVal != nil {
		t.Errorf("expected nil, got %v", anyVal)
	}
}

func TestConvertToWithJSONSchema_Fallbacks(t *testing.T) {
	// Go int is not directly bypass-safe to float64, but should convert correctly via fallback
	val, err := ConvertToWithJSONSchema[int, float64](42, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42.0 {
		t.Errorf("expected 42.0, got %v", val)
	}

	// Custom struct should marshal/unmarshal correctly
	type Source struct {
		Name string `json:"name"`
	}
	type Dest struct {
		Name string `json:"name"`
	}
	src := Source{Name: "Bolt"}
	dest, err := ConvertToWithJSONSchema[Source, Dest](src, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.Name != "Bolt" {
		t.Errorf("expected 'Bolt', got %q", dest.Name)
	}
}

func BenchmarkConvertToWithJSONSchema_String(b *testing.B) {
	val := "hello world"
	for i := 0; i < b.N; i++ {
		_, err := ConvertToWithJSONSchema[string, string](val, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertToWithJSONSchema_Float64(b *testing.B) {
	val := 3.14
	for i := 0; i < b.N; i++ {
		_, err := ConvertToWithJSONSchema[float64, float64](val, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertToWithJSONSchema_Bool(b *testing.B) {
	val := true
	for i := 0; i < b.N; i++ {
		_, err := ConvertToWithJSONSchema[bool, bool](val, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
