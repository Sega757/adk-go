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

	"github.com/google/go-cmp/cmp"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/internal/utils"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

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
			name: "zero value struct",
			part: &genai.Part{},
			want: true,
		},
		{
			name: "with text",
			part: &genai.Part{Text: "hello"},
			want: false,
		},
		{
			name: "with thought true",
			part: &genai.Part{Thought: true},
			want: false,
		},
		{
			name: "with thought signature",
			part: &genai.Part{ThoughtSignature: []byte("sig")},
			want: false,
		},
		{
			name: "with function call",
			part: &genai.Part{FunctionCall: &genai.FunctionCall{Name: "foo"}},
			want: false,
		},
		{
			name: "with function response",
			part: &genai.Part{FunctionResponse: &genai.FunctionResponse{Name: "foo"}},
			want: false,
		},
		{
			name: "with inline data",
			part: &genai.Part{InlineData: &genai.Blob{MIMEType: "image/png"}},
			want: false,
		},
		{
			name: "with file data",
			part: &genai.Part{FileData: &genai.FileData{FileURI: "gs://bucket/file"}},
			want: false,
		},
		{
			name: "with executable code",
			part: &genai.Part{ExecutableCode: &genai.ExecutableCode{Code: "print(1)"}},
			want: false,
		},
		{
			name: "with code execution result",
			part: &genai.Part{CodeExecutionResult: &genai.CodeExecutionResult{Outcome: genai.OutcomeOK}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsZeroPart(tt.part); got != tt.want {
				t.Errorf("IsZeroPart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionCalls(t *testing.T) {
	fc1 := &genai.FunctionCall{Name: "func1"}
	fc2 := &genai.FunctionCall{Name: "func2"}

	tests := []struct {
		name    string
		content *genai.Content
		want    []*genai.FunctionCall
	}{
		{
			name:    "nil content",
			content: nil,
			want:    nil,
		},
		{
			name: "no function calls",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "hello"},
				},
			},
			want: nil,
		},
		{
			name: "extracts function calls",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "thinking..."},
					{FunctionCall: fc1},
					{FunctionCall: fc2},
				},
			},
			want: []*genai.FunctionCall{fc1, fc2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.FunctionCalls(tt.content)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("FunctionCalls() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	fr1 := &genai.FunctionResponse{Name: "func1"}

	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "hello"},
			{FunctionResponse: fr1},
		},
	}

	t.Run("FunctionResponses", func(t *testing.T) {
		got := utils.FunctionResponses(content)
		want := []*genai.FunctionResponse{fr1}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("FunctionResponses() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("TextParts", func(t *testing.T) {
		got := utils.TextParts(content)
		want := []string{"hello"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("TextParts() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Content", func(t *testing.T) {
		event := &session.Event{
			LLMResponse: model.LLMResponse{
				Content: content,
			},
		}
		if got := utils.Content(event); got != content {
			t.Errorf("Content() = %v, want %v", got, content)
		}
		if got := utils.Content(nil); got != nil {
			t.Errorf("Content(nil) = %v, want nil", got)
		}
	})
}

func TestHasFunctionCalls(t *testing.T) {
	contentWithCalls := &genai.Content{
		Parts: []*genai.Part{
			{Text: "thinking..."},
			{FunctionCall: &genai.FunctionCall{Name: "search"}},
		},
	}
	contentWithoutCalls := &genai.Content{
		Parts: []*genai.Part{
			{Text: "hello"},
		},
	}
	partsWithCalls := contentWithCalls.Parts
	partsWithoutCalls := contentWithoutCalls.Parts
	eventWithCalls := &session.Event{LLMResponse: model.LLMResponse{Content: contentWithCalls}}
	eventWithoutCalls := &session.Event{LLMResponse: model.LLMResponse{Content: contentWithoutCalls}}

	tests := []struct {
		name string
		fn   func() bool
		want bool
	}{
		{"Content with calls", func() bool { return utils.HasFunctionCalls(contentWithCalls) }, true},
		{"Content without calls", func() bool { return utils.HasFunctionCalls(contentWithoutCalls) }, false},
		{"Nil content", func() bool { return utils.HasFunctionCalls((*genai.Content)(nil)) }, false},
		{"Parts with calls", func() bool { return utils.HasFunctionCalls(partsWithCalls) }, true},
		{"Parts without calls", func() bool { return utils.HasFunctionCalls(partsWithoutCalls) }, false},
		{"Nil parts", func() bool { return utils.HasFunctionCalls(([]*genai.Part)(nil)) }, false},
		{"Event with calls", func() bool { return utils.HasFunctionCalls(eventWithCalls) }, true},
		{"Event without calls", func() bool { return utils.HasFunctionCalls(eventWithoutCalls) }, false},
		{"Nil event", func() bool { return utils.HasFunctionCalls((*session.Event)(nil)) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
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
			name: "content without function responses",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "hello"},
					{FunctionCall: &genai.FunctionCall{Name: "search"}},
				},
			},
			want: false,
		},
		{
			name: "content with function response",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "result"},
					{FunctionResponse: &genai.FunctionResponse{Name: "search"}},
				},
			},
			want: true,
		},
		{
			name: "content with nil part",
			content: &genai.Content{
				Parts: []*genai.Part{
					nil,
					{FunctionResponse: &genai.FunctionResponse{Name: "search"}},
				},
			},
			want: true,
		},
		{
			name: "content with only nil parts",
			content: &genai.Content{
				Parts: []*genai.Part{nil, nil},
			},
			want: false,
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

func BenchmarkIsZeroPart(b *testing.B) {
	part := &genai.Part{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.IsZeroPart(part)
	}
}

func BenchmarkHasFunctionCalls(b *testing.B) {
	parts := []*genai.Part{
		{Text: "thinking..."},
		{FunctionCall: &genai.FunctionCall{Name: "search"}},
	}
	content := &genai.Content{Parts: parts}
	event := &session.Event{
		LLMResponse: model.LLMResponse{
			Content: content,
		},
	}

	b.Run("Content", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = utils.HasFunctionCalls(content)
		}
	})

	b.Run("Parts", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = utils.HasFunctionCalls(parts)
		}
	})

	b.Run("Event", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = utils.HasFunctionCalls(event)
		}
	})
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
	parts := []*genai.Part{
		{FunctionResponse: &genai.FunctionResponse{Name: "search"}},
	}
	content := &genai.Content{Parts: parts}
	event := &session.Event{
		LLMResponse: model.LLMResponse{
			Content: content,
		},
	}

	b.Run("Content", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = utils.HasFunctionResponses(content)
		}
	})

	b.Run("Parts", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = utils.HasFunctionResponses(parts)
		}
	})

	b.Run("Event", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = utils.HasFunctionResponses(event)
		}
	})
}
