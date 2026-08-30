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

package converters

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type sampleStruct struct {
	Name    string         `json:"name"`
	Age     int            `json:"age,omitempty"`
	Ignored string         `json:"-"`
	Nested  *nestedStruct  `json:"nested,omitempty"`
	Tags    []string       `json:"tags,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

type nestedStruct struct {
	Key string `json:"key"`
}

type recursiveStruct struct {
	Self *recursiveStruct `json:"self"`
}

func TestToMapStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "simple struct",
			input: sampleStruct{
				Name:    "Alice",
				Age:     30,
				Ignored: "secret",
				Nested: &nestedStruct{
					Key: "value",
				},
				Tags: []string{"admin", "user"},
				Meta: map[string]any{
					"role": "owner",
				},
			},
			want: map[string]any{
				"name": "Alice",
				"age":  float64(30), // JSON numbers unmarshal as float64 into map[string]any
				"nested": map[string]any{
					"key": "value",
				},
				"tags": []any{"admin", "user"},
				"meta": map[string]any{
					"role": "owner",
				},
			},
			wantErr: false,
		},
		{
			name: "struct with omitempty zero values",
			input: sampleStruct{
				Name:    "Bob",
				Ignored: "hidden",
			},
			want: map[string]any{
				"name": "Bob",
			},
			wantErr: false,
		},
		{
			name:    "empty struct",
			input:   struct{}{},
			want:    map[string]any{},
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   map[string]any{},
			want:    map[string]any{},
			wantErr: false,
		},
		{
			name:    "untyped nil",
			input:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "typed nil pointer",
			input:   (*sampleStruct)(nil),
			want:    nil,
			wantErr: false,
		},
		{
			name:    "error on unmarshalable type - channel",
			input:   make(chan int),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error on unmarshalable type - function",
			input:   func() {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error on circular reference",
			input: func() any {
				r := &recursiveStruct{}
				r.Self = r
				return r
			}(),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error on primitive input - string",
			input:   "hello",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error on primitive input - int",
			input:   123,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error on slice input",
			input:   []string{"a", "b"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToMapStructure(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ToMapStructure() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToMapStructure() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFromMapStructure(t *testing.T) {
	t.Parallel()

	t.Run("convert to sampleStruct", func(t *testing.T) {
		t.Parallel()

		input := map[string]any{
			"name": "Charlie",
			"age":  25,
			"nested": map[string]any{
				"key": "sub_value",
			},
			"tags": []any{"dev"},
		}

		want := &sampleStruct{
			Name: "Charlie",
			Age:  25,
			Nested: &nestedStruct{
				Key: "sub_value",
			},
			Tags: []string{"dev"},
		}

		got, err := FromMapStructure[sampleStruct](input)
		if err != nil {
			t.Fatalf("FromMapStructure() unexpected error: %v", err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("FromMapStructure() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("convert from nil map", func(t *testing.T) {
		t.Parallel()

		got, err := FromMapStructure[sampleStruct](nil)
		if err != nil {
			t.Fatalf("FromMapStructure() unexpected error: %v", err)
		}

		want := &sampleStruct{}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("FromMapStructure() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("convert from empty map", func(t *testing.T) {
		t.Parallel()

		got, err := FromMapStructure[sampleStruct](map[string]any{})
		if err != nil {
			t.Fatalf("FromMapStructure() unexpected error: %v", err)
		}

		want := &sampleStruct{}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("FromMapStructure() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error on type mismatch", func(t *testing.T) {
		t.Parallel()

		input := map[string]any{
			"name": 12345, // invalid type for string field
		}

		_, err := FromMapStructure[sampleStruct](input)
		if err == nil {
			t.Fatal("FromMapStructure() expected error for type mismatch, got nil")
		}
	})

	t.Run("error when marshalling unmarshalable data inside map", func(t *testing.T) {
		t.Parallel()

		input := map[string]any{
			"fn": func() {}, // JSON marshal fails on functions
		}

		_, err := FromMapStructure[sampleStruct](input)
		if err == nil {
			t.Fatal("FromMapStructure() expected error for unmarshalable map value, got nil")
		}
	})
}

func TestMapStructureRoundTrip(t *testing.T) {
	t.Parallel()

	original := sampleStruct{
		Name: "RoundTripUser",
		Age:  40,
		Nested: &nestedStruct{
			Key: "rt_key",
		},
		Tags: []string{"a", "b"},
		Meta: map[string]any{
			"active": true,
		},
	}

	mapStructure, err := ToMapStructure(original)
	if err != nil {
		t.Fatalf("ToMapStructure() unexpected error: %v", err)
	}

	restored, err := FromMapStructure[sampleStruct](mapStructure)
	if err != nil {
		t.Fatalf("FromMapStructure() unexpected error: %v", err)
	}

	// Expect Ignored field ("secret") to be cleared because of `json:"-"`
	want := original
	want.Ignored = ""

	if diff := cmp.Diff(&want, restored); diff != "" {
		t.Errorf("RoundTrip mismatch (-want +got):\n%s", diff)
	}
}
