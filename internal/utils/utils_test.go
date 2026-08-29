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

package utils_test

import (
	"strings"
	"testing"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/internal/utils"
	"google.golang.org/adk/v2/platform"
)

func TestGenerateFunctionCallIDUsesProvider(t *testing.T) {
	ctx := platform.WithUUIDProvider(t.Context(), func() string { return "fixed" })

	got := utils.GenerateFunctionCallID(ctx)

	// The generated ID must carry the "adk-" prefix that RemoveClientFunctionCallID
	// relies on, and must incorporate the value from the installed provider.
	if !strings.HasPrefix(got, "adk-") {
		t.Errorf("GenerateFunctionCallID() = %q, want \"adk-\" prefix", got)
	}
	if !strings.HasSuffix(got, "fixed") {
		t.Errorf("GenerateFunctionCallID() = %q, want it to use the provider value %q", got, "fixed")
	}
}

func TestGenerateFunctionCallIDDefaultIsUnique(t *testing.T) {
	first := utils.GenerateFunctionCallID(t.Context())
	second := utils.GenerateFunctionCallID(t.Context())

	if first == second {
		t.Errorf("GenerateFunctionCallID() returned %q twice; want unique values", first)
	}
}

func TestPopulateClientFunctionCallIDUsesProvider(t *testing.T) {
	ctx := platform.WithUUIDProvider(t.Context(), func() string { return "generated" })

	content := &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{Name: "needs_id"}},
			{FunctionCall: &genai.FunctionCall{ID: "keep", Name: "has_id"}},
		},
	}

	utils.PopulateClientFunctionCallID(ctx, content)

	if got := content.Parts[0].FunctionCall.ID; got != "adk-generated" {
		t.Errorf("empty function call ID = %q, want %q", got, "adk-generated")
	}
	if got := content.Parts[1].FunctionCall.ID; got != "keep" {
		t.Errorf("preset function call ID = %q, want it left untouched (%q)", got, "keep")
	}
}

func TestRemoveClientFunctionCallID(t *testing.T) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{ID: "adk-12345", Name: "call1"}},
			{FunctionCall: &genai.FunctionCall{ID: "custom-id", Name: "call2"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "adk-67890", Name: "resp1"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "custom-id-2", Name: "resp2"}},
		},
	}

	utils.RemoveClientFunctionCallID(content)

	if got := content.Parts[0].FunctionCall.ID; got != "" {
		t.Errorf("adk function call ID = %q, want empty string", got)
	}
	if got := content.Parts[1].FunctionCall.ID; got != "custom-id" {
		t.Errorf("custom function call ID = %q, want custom-id", got)
	}
	if got := content.Parts[2].FunctionResponse.ID; got != "" {
		t.Errorf("adk function response ID = %q, want empty string", got)
	}
	if got := content.Parts[3].FunctionResponse.ID; got != "custom-id-2" {
		t.Errorf("custom function response ID = %q, want custom-id-2", got)
	}
}

func TestHelperFunctions(t *testing.T) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "hello"},
			{FunctionCall: &genai.FunctionCall{Name: "fn1"}},
			{FunctionResponse: &genai.FunctionResponse{Name: "res1"}},
			{Text: "world"},
		},
	}

	calls := utils.FunctionCalls(content)
	if len(calls) != 1 || calls[0].Name != "fn1" {
		t.Errorf("FunctionCalls = %v, want 1 call named fn1", calls)
	}

	resps := utils.FunctionResponses(content)
	if len(resps) != 1 || resps[0].Name != "res1" {
		t.Errorf("FunctionResponses = %v, want 1 response named res1", resps)
	}

	texts := utils.TextParts(content)
	if len(texts) != 2 || texts[0] != "hello" || texts[1] != "world" {
		t.Errorf("TextParts = %v, want [hello, world]", texts)
	}

	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{Name: "decl1"},
				},
			},
		},
	}
	decls := utils.FunctionDecls(config)
	if len(decls) != 1 || decls[0].Name != "decl1" {
		t.Errorf("FunctionDecls = %v, want 1 decl named decl1", decls)
	}
}

func BenchmarkPopulateClientFunctionCallID(b *testing.B) {
	ctx := platform.WithUUIDProvider(b.Context(), func() string { return "bench-uuid" })

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		content := &genai.Content{
			Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{Name: "needs_id_1"}},
				{FunctionCall: &genai.FunctionCall{ID: "existing_id", Name: "has_id"}},
				{FunctionCall: &genai.FunctionCall{Name: "needs_id_2"}},
			},
		}
		utils.PopulateClientFunctionCallID(ctx, content)
	}
}

func BenchmarkRemoveClientFunctionCallID(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		content := &genai.Content{
			Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{ID: "adk-12345", Name: "call1"}},
				{FunctionCall: &genai.FunctionCall{ID: "custom-id", Name: "call2"}},
				{FunctionResponse: &genai.FunctionResponse{ID: "adk-67890", Name: "resp1"}},
			},
		}
		utils.RemoveClientFunctionCallID(content)
	}
}
