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
)

func TestAppendInstructions(t *testing.T) {
	tests := []struct {
		name         string
		initialReq   func() *model.LLMRequest
		instructions []string
		want         *model.LLMRequest
	}{
		{
			name: "no instructions leaves request unchanged",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{}
			},
			instructions: nil,
			want:         &model.LLMRequest{},
		},
		{
			name: "no instructions with empty slice leaves request unchanged",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{}
			},
			instructions: []string{},
			want:         &model.LLMRequest{},
		},
		{
			name: "nil config initializes config and system instruction",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{}
			},
			instructions: []string{"System Prompt 1"},
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{Text: "System Prompt 1"},
						},
					},
				},
			},
		},
		{
			name: "nil system instruction initializes system instruction on existing config while preserving config fields",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{
					Config: &genai.GenerateContentConfig{
						Temperature: genai.Ptr(float32(0.7)),
					},
				}
			},
			instructions: []string{"System Prompt 1"},
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					Temperature: genai.Ptr(float32(0.7)),
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{Text: "System Prompt 1"},
						},
					},
				},
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
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{Text: "Initial instruction\n\nAdditional instruction"},
						},
					},
				},
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
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{Text: "New instruction"},
						},
					},
				},
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
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{Text: ""},
							{Text: "New instruction"},
						},
					},
				},
			},
		},
		{
			name: "existing system instruction with non-text last part appends new text part",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{
					Config: &genai.GenerateContentConfig{
						SystemInstruction: &genai.Content{
							Role: genai.RoleUser,
							Parts: []*genai.Part{
								{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte("fake")}},
							},
						},
					},
				}
			},
			instructions: []string{"Instruction following inline data"},
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte("fake")}},
							{Text: "Instruction following inline data"},
						},
					},
				},
			},
		},
		{
			name: "multiple variadic instructions joined with double newline",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{}
			},
			instructions: []string{"Inst 1", "Inst 2", "Inst 3"},
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{Text: "Inst 1\n\nInst 2\n\nInst 3"},
						},
					},
				},
			},
		},
		{
			name: "empty string instruction appended cleanly",
			initialReq: func() *model.LLMRequest {
				return &model.LLMRequest{
					Config: &genai.GenerateContentConfig{
						SystemInstruction: genai.NewContentFromText("Header", genai.RoleUser),
					},
				}
			},
			instructions: []string{""},
			want: &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Role: genai.RoleUser,
						Parts: []*genai.Part{
							{Text: "Header\n\n"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.initialReq()
			utils.AppendInstructions(got, tt.instructions...)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AppendInstructions() mismatch (-want +got):\n%s", diff)
			}
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
