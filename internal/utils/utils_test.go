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
	"google.golang.org/adk/v2/model"
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

func TestIsZeroPart(t *testing.T) {
	tests := []struct {
		name string
		part *genai.Part
		want bool
	}{
		{
			name: "nil part",
			part: nil,
			want: true,
		},
		{
			name: "empty struct part",
			part: &genai.Part{},
			want: true,
		},
		{
			name: "text part",
			part: &genai.Part{Text: "hello"},
			want: false,
		},
		{
			name: "thought part",
			part: &genai.Part{Thought: true},
			want: false,
		},
		{
			name: "thought signature part",
			part: &genai.Part{ThoughtSignature: []byte("sig")},
			want: false,
		},
		{
			name: "function call part",
			part: &genai.Part{FunctionCall: &genai.FunctionCall{Name: "fn"}},
			want: false,
		},
		{
			name: "part metadata part",
			part: &genai.Part{PartMetadata: map[string]any{"key": "val"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsZeroPart(tt.part)
			if got != tt.want {
				t.Errorf("IsZeroPart(%v) = %v, want %v", tt.part, got, tt.want)
			}
		})
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

func TestHasFunctionCalls(t *testing.T) {
	tests := []struct {
		name    string
		content *genai.Content
		want    bool
	}{
		{
			name:    "nil content",
			content: nil,
			want:    false,
		},
		{
			name:    "empty parts",
			content: &genai.Content{Parts: []*genai.Part{}},
			want:    false,
		},
		{
			name:    "text only parts",
			content: &genai.Content{Parts: []*genai.Part{{Text: "hello"}}},
			want:    false,
		},
		{
			name:    "function call present",
			content: &genai.Content{Parts: []*genai.Part{{Text: "call"}, {FunctionCall: &genai.FunctionCall{Name: "fn"}}}},
			want:    true,
		},
		{
			name:    "function response present, no function call",
			content: &genai.Content{Parts: []*genai.Part{{FunctionResponse: &genai.FunctionResponse{Name: "fn"}}}},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.HasFunctionCalls(tt.content); got != tt.want {
				t.Errorf("HasFunctionCalls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasFunctionResponses(t *testing.T) {
	tests := []struct {
		name    string
		content *genai.Content
		want    bool
	}{
		{
			name:    "nil content",
			content: nil,
			want:    false,
		},
		{
			name:    "empty parts",
			content: &genai.Content{Parts: []*genai.Part{}},
			want:    false,
		},
		{
			name:    "text only parts",
			content: &genai.Content{Parts: []*genai.Part{{Text: "hello"}}},
			want:    false,
		},
		{
			name:    "function response present",
			content: &genai.Content{Parts: []*genai.Part{{Text: "resp"}, {FunctionResponse: &genai.FunctionResponse{Name: "fn"}}}},
			want:    true,
		},
		{
			name:    "function call present, no function response",
			content: &genai.Content{Parts: []*genai.Part{{FunctionCall: &genai.FunctionCall{Name: "fn"}}}},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.HasFunctionResponses(tt.content); got != tt.want {
				t.Errorf("HasFunctionResponses() = %v, want %v", got, tt.want)
			}
		})
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

func BenchmarkIsZeroPart(b *testing.B) {
	part := &genai.Part{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.IsZeroPart(part)
	}
}

func BenchmarkHasFunctionCalls(b *testing.B) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "thinking..."},
			{FunctionCall: &genai.FunctionCall{Name: "search"}},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.HasFunctionCalls(content)
	}
}

func BenchmarkHasFunctionCalls_LegacyFunctionCallsSlice(b *testing.B) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "thinking..."},
			{FunctionCall: &genai.FunctionCall{Name: "search"}},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(utils.FunctionCalls(content)) > 0
	}
}

func BenchmarkHasFunctionResponses(b *testing.B) {
	content := &genai.Content{
		Parts: []*genai.Part{
			{FunctionResponse: &genai.FunctionResponse{Name: "search"}},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.HasFunctionResponses(content)
	}
}

func TestAppendInstructions(t *testing.T) {
	tests := []struct {
		name         string
		initialReq   func() *model.LLMRequest
		instructions []string
		want         func(*testing.T, *model.LLMRequest)
	}{
		{
			name: "no instructions leaves request unchanged",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{}
			},
			instructions: nil,
			want: func(t *testing.T, r *model.LLMRequest) {
				if r.Config != nil {
					t.Errorf("expected Config to remain nil, got %v", r.Config)
				}
			},
		},
		{
			name: "nil config initializes config and system instruction",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{}
			},
			instructions: []string{"System Prompt 1"},
			want: func(t *testing.T, r *model.LLMRequest) {
				if r.Config == nil {
					t.Fatal("expected Config to be initialized, got nil")
				}
				if r.Config.SystemInstruction == nil {
					t.Fatal("expected SystemInstruction to be initialized, got nil")
				}
				if len(r.Config.SystemInstruction.Parts) != 1 {
					t.Fatalf("expected 1 part, got %d", len(r.Config.SystemInstruction.Parts))
				}
				if got := r.Config.SystemInstruction.Parts[0].Text; got != "System Prompt 1" {
					t.Errorf("part text = %q, want %q", got, "System Prompt 1")
				}
				if got := r.Config.SystemInstruction.Role; got != genai.RoleUser {
					t.Errorf("role = %q, want %q", got, genai.RoleUser)
				}
			},
		},
		{
			name: "nil system instruction initializes system instruction on existing config",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{
					Config: &genai.GenerateContentConfig{
						Temperature: genai.Ptr(float32(0.7)),
					},
				}
			},
			instructions: []string{"System Prompt 1"},
			want: func(t *testing.T, r *model.LLMRequest) {
				if r.Config == nil || r.Config.SystemInstruction == nil {
					t.Fatal("expected SystemInstruction to be initialized")
				}
				if got := *r.Config.Temperature; got != float32(0.7) {
					t.Errorf("expected Temperature to be preserved, got %f", got)
				}
				if got := r.Config.SystemInstruction.Parts[0].Text; got != "System Prompt 1" {
					t.Errorf("part text = %q, want %q", got, "System Prompt 1")
				}
			},
		},
		{
			name: "existing system instruction appends to last non-empty text part",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{
					Config: &genai.GenerateContentConfig{
						SystemInstruction: genai.NewContentFromText("Initial instruction", genai.RoleUser),
					},
				}
			},
			instructions: []string{"Additional instruction"},
			want: func(t *testing.T, r *model.LLMRequest) {
				if r.Config == nil || r.Config.SystemInstruction == nil {
					t.Fatal("expected SystemInstruction to exist")
				}
				if len(r.Config.SystemInstruction.Parts) != 1 {
					t.Fatalf("expected 1 part, got %d", len(r.Config.SystemInstruction.Parts))
				}
				wantText := "Initial instruction\n\nAdditional instruction"
				if got := r.Config.SystemInstruction.Parts[0].Text; got != wantText {
					t.Errorf("part text = %q, want %q", got, wantText)
				}
			},
		},
		{
			name: "existing system instruction with empty parts appends new part",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{
					Config: &genai.GenerateContentConfig{
						SystemInstruction: &genai.Content{
							Role:  genai.RoleUser,
							Parts: []*genai.Part{},
						},
					},
				}
			},
			instructions: []string{"New instruction"},
			want: func(t *testing.T, r *model.LLMRequest) {
				if r.Config == nil || r.Config.SystemInstruction == nil {
					t.Fatal("expected SystemInstruction to exist")
				}
				if len(r.Config.SystemInstruction.Parts) != 1 {
					t.Fatalf("expected 1 part, got %d", len(r.Config.SystemInstruction.Parts))
				}
				if got := r.Config.SystemInstruction.Parts[0].Text; got != "New instruction" {
					t.Errorf("part text = %q, want %q", got, "New instruction")
				}
			},
		},
		{
			name: "existing system instruction with empty last part text appends new part",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{
					Config: &genai.GenerateContentConfig{
						SystemInstruction: &genai.Content{
							Role: genai.RoleUser,
							Parts: []*genai.Part{
								{Text: ""},
							},
						},
					},
				}
			},
			instructions: []string{"New instruction"},
			want: func(t *testing.T, r *model.LLMRequest) {
				if r.Config == nil || r.Config.SystemInstruction == nil {
					t.Fatal("expected SystemInstruction to exist")
				}
				if len(r.Config.SystemInstruction.Parts) != 2 {
					t.Fatalf("expected 2 parts, got %d", len(r.Config.SystemInstruction.Parts))
				}
				if got := r.Config.SystemInstruction.Parts[1].Text; got != "New instruction" {
					t.Errorf("second part text = %q, want %q", got, "New instruction")
				}
			},
		},
		{
			name: "multiple variadic instructions joined with double newline",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{}
			},
			instructions: []string{"Inst 1", "Inst 2", "Inst 3"},
			want: func(t *testing.T, r *model.LLMRequest) {
				if r.Config == nil || r.Config.SystemInstruction == nil {
					t.Fatal("expected SystemInstruction to exist")
				}
				wantText := "Inst 1\n\nInst 2\n\nInst 3"
				if got := r.Config.SystemInstruction.Parts[0].Text; got != wantText {
					t.Errorf("part text = %q, want %q", got, wantText)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.initialReq()
			utils.AppendInstructions(req, tt.instructions...)
			tt.want(t, req)
		})
	}
}

func TestAppendInstructionsSequential(t *testing.T) {
	req := &model.LLMRequest{}

	utils.AppendInstructions(req, "First")
	utils.AppendInstructions(req, "Second")
	utils.AppendInstructions(req, "Third", "Fourth")

	if req.Config == nil || req.Config.SystemInstruction == nil {
		t.Fatal("expected SystemInstruction to exist")
	}

	wantText := "First\n\nSecond\n\nThird\n\nFourth"
	if got := req.Config.SystemInstruction.Parts[0].Text; got != wantText {
		t.Errorf("part text = %q, want %q", got, wantText)
	}
}

func BenchmarkAppendInstructions(b *testing.B) {
	b.Run("single append", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: genai.NewContentFromText("Initial", genai.RoleUser),
				},
			}
			utils.AppendInstructions(req, "New instruction")
		}
	})

	b.Run("variadic append", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := &model.LLMRequest{}
			utils.AppendInstructions(req, "First instruction", "Second instruction", "Third instruction")
		}
	})
}
