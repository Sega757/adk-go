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
	// Test basic conversions
	res, err := ConvertToWithJSONSchema[string, string]("hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "hello" {
		t.Errorf("expected hello, got %s", res)
	}

	resFloat, err := ConvertToWithJSONSchema[float64, any](3.14, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resFloat != 3.14 {
		t.Errorf("expected 3.14, got %v", resFloat)
	}

	// Ensure ints (which are not in our isJSONSafe check to avoid float64 mixups)
	// still convert fine through fallback json marshal/unmarshal.
	resInt, err := ConvertToWithJSONSchema[int, float64](42, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resInt != 42.0 {
		t.Errorf("expected 42.0, got %v", resInt)
	}
}

func BenchmarkConvertToWithJSONSchema_String(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[string, string]("hello", nil)
	}
}

func BenchmarkConvertToWithJSONSchema_Float64(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToWithJSONSchema[float64, any](3.14, nil)
	}
}
