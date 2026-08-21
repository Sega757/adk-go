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
			{FunctionCall: &genai.FunctionCall{ID: "adk-1234", Name: "test_call"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "adk-5678", Name: "test_resp"}},
			{FunctionCall: &genai.FunctionCall{ID: "custom-id", Name: "keep_call"}},
		},
	}

	utils.RemoveClientFunctionCallID(content)

	if got := content.Parts[0].FunctionCall.ID; got != "" {
		t.Errorf("expected empty ID for adk- call, got %q", got)
	}
	if got := content.Parts[1].FunctionResponse.ID; got != "" {
		t.Errorf("expected empty ID for adk- response, got %q", got)
	}
	if got := content.Parts[2].FunctionCall.ID; got != "custom-id" {
		t.Errorf("expected custom-id preserved, got %q", got)
	}
}

func BenchmarkPopulateClientFunctionCallID(b *testing.B) {
	ctx := platform.WithUUIDProvider(b.Context(), func() string { return "fixed" })
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "hello"},
			{FunctionCall: &genai.FunctionCall{ID: "already_set", Name: "foo"}},
			{Text: "world"},
			{FunctionCall: &genai.FunctionCall{ID: "also_set", Name: "bar"}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		utils.PopulateClientFunctionCallID(ctx, content)
	}
}

func BenchmarkRemoveClientFunctionCallID(b *testing.B) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "hello"},
			{FunctionCall: &genai.FunctionCall{ID: "custom-id", Name: "foo"}},
			{Text: "world"},
			{FunctionResponse: &genai.FunctionResponse{ID: "custom-resp", Name: "bar"}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		utils.RemoveClientFunctionCallID(content)
	}
}

func BenchmarkFunctionCalls(b *testing.B) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "intro"},
			{FunctionCall: &genai.FunctionCall{Name: "fn1"}},
			{Text: "middle"},
			{FunctionCall: &genai.FunctionCall{Name: "fn2"}},
			{FunctionCall: &genai.FunctionCall{Name: "fn3"}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = utils.FunctionCalls(content)
	}
}
