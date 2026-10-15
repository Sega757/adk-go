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
	"testing"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/internal/utils"
	"google.golang.org/adk/v2/platform"
)

func TestGenerateFunctionCallIDUsesProvider(t *testing.T) {
	ctx := platform.WithUUIDProvider(t.Context(), func() string { return "fixed" })

	got := utils.GenerateFunctionCallID(ctx)

	if got != "adk-fixed" {
		t.Errorf("GenerateFunctionCallID() = %q, want %q", got, "adk-fixed")
	}
}

func TestGenerateFunctionCallIDDefaultIsUnique(t *testing.T) {
	ctx := t.Context()

	got1 := utils.GenerateFunctionCallID(ctx)
	got2 := utils.GenerateFunctionCallID(ctx)

	if got1 == got2 {
		t.Errorf("GenerateFunctionCallID() generated duplicate IDs: %q", got1)
	}
}

func TestPopulateClientFunctionCallIDUsesProvider(t *testing.T) {
	ctx := platform.WithUUIDProvider(t.Context(), func() string { return "fixed" })
	content := &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{Name: "fn1"}},
			{FunctionCall: &genai.FunctionCall{ID: "keep", Name: "fn2"}},
		},
	}

	utils.PopulateClientFunctionCallID(ctx, content)

	if got := content.Parts[0].FunctionCall.ID; got != "adk-fixed" {
		t.Errorf("first FunctionCall.ID = %q, want %q", got, "adk-fixed")
	}
	if got := content.Parts[1].FunctionCall.ID; got != "keep" {
		t.Errorf("second FunctionCall.ID = %q, want %q", got, "keep")
	}
}

func TestRemoveClientFunctionCallID(t *testing.T) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{ID: "adk-fixed", Name: "fn1"}},
			{FunctionCall: &genai.FunctionCall{ID: "keep", Name: "fn2"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "adk-fixed", Name: "fn1"}},
			{FunctionResponse: &genai.FunctionResponse{ID: "keep", Name: "fn2"}},
		},
	}

	utils.RemoveClientFunctionCallID(content)

	if got := content.Parts[0].FunctionCall.ID; got != "" {
		t.Errorf("first FunctionCall.ID = %q, want empty", got)
	}
	if got := content.Parts[1].FunctionCall.ID; got != "keep" {
		t.Errorf("second FunctionCall.ID = %q, want %q", got, "keep")
	}
	if got := content.Parts[2].FunctionResponse.ID; got != "" {
		t.Errorf("first FunctionResponse.ID = %q, want empty", got)
	}
	if got := content.Parts[3].FunctionResponse.ID; got != "keep" {
		t.Errorf("second FunctionResponse.ID = %q, want %q", got, "keep")
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
