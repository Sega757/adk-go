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

package utils

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func BenchmarkCanonicalSchemaJSON(b *testing.B) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"foo": {Type: "string"},
			"bar": {Type: "integer"},
			"nested": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"a": {Type: "string"},
					"b": {Type: "boolean"},
				},
			},
		},
		PropertyOrder: []string{"foo", "bar", "nested"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CanonicalSchemaJSON(schema)
		if err != nil {
			b.Fatalf("failed to canonicalize: %v", err)
		}
	}
}

func BenchmarkCanonicalize(b *testing.B) {
	input := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"foo": map[string]any{"type": "string"},
			"bar": map[string]any{"type": "integer"},
			"nested": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"a": map[string]any{"type": "string"},
					"b": map[string]any{"type": "boolean"},
				},
			},
		},
		"required": []any{"foo", "nested"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := canonicalize(input)
		if err != nil {
			b.Fatalf("failed to canonicalize: %v", err)
		}
	}
}
