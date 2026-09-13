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

func TestCloneGenerateContentConfig(t *testing.T) {
	temp := float32(0.7)
	topP := float32(0.9)
	topK := float32(40)
	logprobs := int32(5)
	presencePenalty := float32(0.1)
	frequencyPenalty := float32(0.2)
	seed := int32(42)
	enableCivic := true

	orig := &genai.GenerateContentConfig{
		Temperature:                &temp,
		TopP:                       &topP,
		TopK:                       &topK,
		CandidateCount:             1,
		MaxOutputTokens:            100,
		StopSequences:              []string{"END", "STOP"},
		ResponseLogprobs:           true,
		Logprobs:                   &logprobs,
		PresencePenalty:            &presencePenalty,
		FrequencyPenalty:           &frequencyPenalty,
		Seed:                       &seed,
		ResponseMIMEType:           "application/json",
		SystemInstruction:          &genai.Content{Role: "system", Parts: []*genai.Part{{Text: "System prompt"}}},
		ResponseSchema:             &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{"key": {Type: genai.TypeString}}},
		ResponseJsonSchema:         map[string]any{"type": "object"},
		RoutingConfig:              &genai.GenerationConfigRoutingConfig{AutoMode: &genai.GenerationConfigRoutingConfigAutoRoutingMode{}},
		ModelSelectionConfig:       &genai.ModelSelectionConfig{},
		SafetySettings:             []*genai.SafetySetting{{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_LOW_AND_ABOVE"}},
		Tools:                      []*genai.Tool{{FunctionDeclarations: []*genai.FunctionDeclaration{{Name: "fn1", Description: "desc1"}}}},
		ToolConfig:                 &genai.ToolConfig{FunctionCallingConfig: &genai.FunctionCallingConfig{Mode: genai.FunctionCallingConfigModeAny, AllowedFunctionNames: []string{"fn1"}}},
		Labels:                     map[string]string{"env": "prod"},
		ResponseModalities:         []string{"TEXT"},
		SpeechConfig:               &genai.SpeechConfig{},
		ThinkingConfig:             &genai.ThinkingConfig{},
		ImageConfig:                &genai.ImageConfig{},
		EnableEnhancedCivicAnswers: &enableCivic,
		ModelArmorConfig:           &genai.ModelArmorConfig{},
	}

	cloned := clone(orig)
	if !reflect.DeepEqual(orig, cloned) {
		t.Errorf("clone(*genai.GenerateContentConfig) mismatch (-want +got):\n%+v", cloned)
	}

	// Verify deep copy independence:
	*cloned.Temperature = 0.1
	cloned.StopSequences[0] = "MODIFIED"
	cloned.Labels["env"] = "dev"
	cloned.Tools[0].FunctionDeclarations[0].Name = "modified_fn"
	cloned.ToolConfig.FunctionCallingConfig.AllowedFunctionNames[0] = "modified_fn"
	cloned.SystemInstruction.Parts[0].Text = "modified_prompt"

	if *orig.Temperature == 0.1 {
		t.Errorf("original temperature was modified")
	}
	if orig.StopSequences[0] == "MODIFIED" {
		t.Errorf("original stop sequences were modified")
	}
	if orig.Labels["env"] == "dev" {
		t.Errorf("original labels map was modified")
	}
	if orig.Tools[0].FunctionDeclarations[0].Name == "modified_fn" {
		t.Errorf("original function declaration name was modified")
	}
	if orig.ToolConfig.FunctionCallingConfig.AllowedFunctionNames[0] == "modified_fn" {
		t.Errorf("original allowed function names were modified")
	}
	if orig.SystemInstruction.Parts[0].Text == "modified_prompt" {
		t.Errorf("original system instruction text was modified")
	}

	// Test value type cloning
	clonedVal := clone(*orig)
	if !reflect.DeepEqual(*orig, clonedVal) {
		t.Errorf("clone(genai.GenerateContentConfig) value mismatch")
	}

	// Test nil pointer cloning
	var nilConfig *genai.GenerateContentConfig
	if clonedNil := clone(nilConfig); clonedNil != nil {
		t.Errorf("clone(nil *genai.GenerateContentConfig) = %v, want nil", clonedNil)
	}
}

func TestCloneGenAISchema(t *testing.T) {
	orig := &genai.Schema{
		Type:        genai.TypeObject,
		Description: "test schema",
		Enum:        []string{"a", "b"},
		Required:    []string{"key1"},
		Properties: map[string]*genai.Schema{
			"key1": {Type: genai.TypeString, Description: "prop1"},
		},
		Items: &genai.Schema{Type: genai.TypeString},
		AnyOf: []*genai.Schema{{Type: genai.TypeInteger}},
	}

	cloned := clone(orig)
	if !reflect.DeepEqual(orig, cloned) {
		t.Errorf("clone(*genai.Schema) mismatch (-want +got):\n%+v", cloned)
	}

	// Verify deep copy independence:
	cloned.Enum[0] = "MODIFIED"
	cloned.Required[0] = "MODIFIED"
	cloned.Properties["key1"].Description = "MODIFIED"

	if orig.Enum[0] == "MODIFIED" {
		t.Errorf("original enum was modified")
	}
	if orig.Required[0] == "MODIFIED" {
		t.Errorf("original required slice was modified")
	}
	if orig.Properties["key1"].Description == "MODIFIED" {
		t.Errorf("original properties schema was modified")
	}

	// Test value type cloning
	clonedVal := clone(*orig)
	if !reflect.DeepEqual(*orig, clonedVal) {
		t.Errorf("clone(genai.Schema) value mismatch")
	}

	// Test nil pointer cloning
	var nilSchema *genai.Schema
	if clonedNil := clone(nilSchema); clonedNil != nil {
		t.Errorf("clone(nil *genai.Schema) = %v, want nil", clonedNil)
	}
}

func TestCloneGenAITool(t *testing.T) {
	orig := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{Name: "fn1", Description: "desc1", Parameters: &genai.Schema{Type: genai.TypeObject}},
		},
		GoogleSearch: &genai.GoogleSearch{},
	}

	cloned := clone(orig)
	if !reflect.DeepEqual(orig, cloned) {
		t.Errorf("clone(*genai.Tool) mismatch (-want +got):\n%+v", cloned)
	}

	cloned.FunctionDeclarations[0].Name = "modified"
	if orig.FunctionDeclarations[0].Name == "modified" {
		t.Errorf("original tool function declaration was modified")
	}

	clonedVal := clone(*orig)
	if !reflect.DeepEqual(*orig, clonedVal) {
		t.Errorf("clone(genai.Tool) value mismatch")
	}
}

func TestCloneGenAIPart(t *testing.T) {
	orig := &genai.Part{
		Text: "hello",
		FunctionCall: &genai.FunctionCall{
			Name: "fn",
			Args: map[string]any{"arg1": "val1"},
		},
	}

	cloned := clone(orig)
	if !reflect.DeepEqual(orig, cloned) {
		t.Errorf("clone(*genai.Part) mismatch (-want +got):\n%+v", cloned)
	}

	cloned.FunctionCall.Args["arg1"] = "modified"
	if orig.FunctionCall.Args["arg1"] == "modified" {
		t.Errorf("original part function call args map was modified")
	}

	clonedVal := clone(*orig)
	if !reflect.DeepEqual(*orig, clonedVal) {
		t.Errorf("clone(genai.Part) value mismatch")
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

func BenchmarkClone_GenAISchema(b *testing.B) {
	orig := &genai.Schema{
		Type:        genai.TypeObject,
		Description: "JSON response schema",
		Properties: map[string]*genai.Schema{
			"name": {Type: genai.TypeString, Description: "user name"},
			"age":  {Type: genai.TypeInteger, Description: "user age"},
			"tags": {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
		},
		Required: []string{"name", "age"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = clone(orig)
	}
}

func BenchmarkClone_GenAITool(b *testing.B) {
	orig := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "get_weather",
				Description: "Get weather information for location",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"city": {Type: genai.TypeString},
					},
					Required: []string{"city"},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = clone(orig)
	}
}

func BenchmarkClone_GenAIPart(b *testing.B) {
	orig := &genai.Part{
		FunctionCall: &genai.FunctionCall{
			ID:   "call_123",
			Name: "search",
			Args: map[string]any{"query": "golang optimization", "limit": 10},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = clone(orig)
	}
}

func BenchmarkClone_GenerateContentConfig(b *testing.B) {
	temp := float32(0.7)
	topP := float32(0.9)
	orig := &genai.GenerateContentConfig{
		Temperature:       &temp,
		TopP:              &topP,
		SystemInstruction: &genai.Content{Role: "system", Parts: []*genai.Part{{Text: "System prompt"}}},
		ResponseMIMEType:  "application/json",
		StopSequences:     []string{"END", "STOP"},
		Labels:            map[string]string{"env": "prod", "tier": "gold"},
		SafetySettings: []*genai.SafetySetting{
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_LOW_AND_ABOVE"},
		},
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{Name: "my_func", Description: "a test function"},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = clone(orig)
	}
}
