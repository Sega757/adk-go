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

package llminternal

import (
	"reflect"
	"testing"

	"google.golang.org/genai"
)

func TestClone(t *testing.T) {
	type testStruct struct {
		S  string
		I  int
		Sl []string
		M  map[string]string
		P  *int
		N  *testStruct
	}

	testData := func() *testStruct {
		return &testStruct{
			S:  "test",
			I:  123,
			Sl: []string{"a", "b"},
			M:  map[string]string{"k": "v"},
			P:  func() *int { i := 456; return &i }(),
			N: &testStruct{
				S: "nested",
			},
		}
	}

	check := func(t *testing.T, original, cloned *testStruct) {
		if !reflect.DeepEqual(original, cloned) {
			t.Errorf("clone() = %+v, want %+v", cloned, original)
		}

		// Modify cloned and check if original is affected
		cloned.Sl[0] = "c"
		cloned.M["k"] = "v2"
		*cloned.P = 789
		cloned.N.S = "nested2"

		if reflect.DeepEqual(original, cloned) {
			t.Errorf("clone() should not be affected by modifications to original")
		}
		if original.Sl[0] != "a" {
			t.Errorf("original slice was modified")
		}
		if original.M["k"] != "v" {
			t.Errorf("original map was modified")
		}
		if *original.P != 456 {
			t.Errorf("original pointer value was modified")
		}
		if original.N.S != "nested" {
			t.Errorf("original nested struct was modified")
		}
	}

	t.Run("pointer", func(t *testing.T) {
		original := testData()
		cloned := clone(original)
		check(t, original, cloned)
	})
	t.Run("value", func(t *testing.T) {
		original := testData()
		cloned := clone(*original)
		check(t, original, &cloned)
	})
	t.Run("interface", func(t *testing.T) {
		original := testData()
		cloned := clone(any(original))
		typed, ok := cloned.(*testStruct)
		if !ok {
			t.Fatalf("clone failed with interface: %v", cloned)
		}
		check(t, original, typed)
	})
}

func TestCloneNil(t *testing.T) {
	var original *int
	cloned := clone(original)
	if cloned != nil {
		t.Errorf("clone(nil) = %v, want nil", cloned)
	}
}

func TestCloneUnexported(t *testing.T) {
	type testStructUnexported struct {
		s string
	}
	original := &testStructUnexported{s: "test"}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("clone() did not panic on unexported field")
		}
	}()
	clone(original)
}

func TestCloneGenAIContent(t *testing.T) {
	orig := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: "hello"},
			{Thought: true, ThoughtSignature: []byte("sig")},
			{FunctionCall: &genai.FunctionCall{ID: "c1", Name: "fn1", Args: map[string]any{"arg1": "val1"}}},
			{FunctionResponse: &genai.FunctionResponse{ID: "c1", Name: "fn1", Response: map[string]any{"res1": "val2"}}},
			{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte("data")}},
			{FileData: &genai.FileData{MIMEType: "image/png", FileURI: "gs://file"}},
			{ExecutableCode: &genai.ExecutableCode{Code: "print(1)", Language: "PYTHON"}},
			{CodeExecutionResult: &genai.CodeExecutionResult{Outcome: "OK", Output: "1"}},
			{MediaResolution: &genai.PartMediaResolution{}},
			{VideoMetadata: &genai.VideoMetadata{}},
			{ToolCall: &genai.ToolCall{ID: "tc1"}},
			{ToolResponse: &genai.ToolResponse{ID: "tr1"}},
			{PartMetadata: map[string]any{"meta": "data"}},
		},
	}

	cloned := clone(orig)
	if !reflect.DeepEqual(orig, cloned) {
		t.Errorf("clone(*genai.Content) mismatch (-want +got):\n%+v", cloned)
	}

	// Verify deep copy independence:
	cloned.Parts[0].Text = "modified"
	cloned.Parts[2].FunctionCall.Args["arg1"] = "modified"
	cloned.Parts[3].FunctionResponse.Response["res1"] = "modified"
	cloned.Parts[4].InlineData.Data[0] = 'X'

	if orig.Parts[0].Text == "modified" {
		t.Errorf("original text part was modified")
	}
	if orig.Parts[2].FunctionCall.Args["arg1"] == "modified" {
		t.Errorf("original function call args were modified")
	}
	if orig.Parts[3].FunctionResponse.Response["res1"] == "modified" {
		t.Errorf("original function response map was modified")
	}
	if orig.Parts[4].InlineData.Data[0] == 'X' {
		t.Errorf("original inline data slice was modified")
	}

	// Test value type cloning
	clonedVal := clone(*orig)
	if !reflect.DeepEqual(*orig, clonedVal) {
		t.Errorf("clone(genai.Content) value mismatch")
	}

	// Test nil pointer cloning
	var nilContent *genai.Content
	if clonedNil := clone(nilContent); clonedNil != nil {
		t.Errorf("clone(nil *genai.Content) = %v, want nil", clonedNil)
	}
}

func BenchmarkClone(b *testing.B) {
	type testStruct struct {
		S  string
		I  int
		Sl []string
		M  map[string]string
		P  *int
		N  *testStruct
	}

	testData := &testStruct{
		S:  "test",
		I:  123,
		Sl: []string{"a", "b", "c", "d"},
		M:  map[string]string{"k1": "v1", "k2": "v2"},
		P:  func() *int { i := 456; return &i }(),
		N: &testStruct{
			S: "nested",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = clone(testData)
	}
}

func BenchmarkClone_GenAIContent(b *testing.B) {
	orig := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: "hello model"},
			{FunctionCall: &genai.FunctionCall{ID: "c1", Name: "fn1", Args: map[string]any{"arg1": "val1"}}},
			{FunctionResponse: &genai.FunctionResponse{ID: "c1", Name: "fn1", Response: map[string]any{"res1": "val2"}}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = clone(orig)
	}
}
