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
			{FunctionCall: &genai.FunctionCall{ID: "adk-1234", Name: "fc1"}},
			{FunctionCall: &genai.FunctionCall{ID: "user-provided", Name: "fc2"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "adk-5678", Name: "fr1"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "keep-me", Name: "fr2"}},
		},
	}

	utils.RemoveClientFunctionCallID(content)

	if got := content.Parts[0].FunctionCall.ID; got != "" {
		t.Errorf("adk function call ID = %q, want empty string", got)
	}
	if got := content.Parts[1].FunctionCall.ID; got != "user-provided" {
		t.Errorf("preset function call ID = %q, want %q", got, "user-provided")
	}
	if got := content.Parts[2].FunctionResponse.ID; got != "" {
		t.Errorf("adk function response ID = %q, want empty string", got)
	}
	if got := content.Parts[3].FunctionResponse.ID; got != "keep-me" {
		t.Errorf("preset function response ID = %q, want %q", got, "keep-me")
	}
}

func BenchmarkPopulateClientFunctionCallID(b *testing.B) {
	ctx := platform.WithUUIDProvider(b.Context(), func() string { return "bench" })
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "some prompt"},
			{FunctionCall: &genai.FunctionCall{Name: "fn1"}},
			{FunctionCall: &genai.FunctionCall{Name: "fn2"}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		content.Parts[1].FunctionCall.ID = ""
		content.Parts[2].FunctionCall.ID = ""
		utils.PopulateClientFunctionCallID(ctx, content)
	}
}

func BenchmarkRemoveClientFunctionCallID(b *testing.B) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "some prompt"},
			{FunctionCall: &genai.FunctionCall{ID: "adk-123", Name: "fn1"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "adk-456", Name: "fn1"}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		content.Parts[1].FunctionCall.ID = "adk-123"
		content.Parts[2].FunctionResponse.ID = "adk-456"
		utils.RemoveClientFunctionCallID(content)
	}
}
