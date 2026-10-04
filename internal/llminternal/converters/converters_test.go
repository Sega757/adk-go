// Copyright 2026 Google LLC
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
	"google.golang.org/genai"

	"google.golang.org/adk/v2/model"
)

func TestGenai2LLMResponse(t *testing.T) {
	t.Parallel()

	testUsage := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount: 10,
		TotalTokenCount:  25,
	}
	testGrounding := &genai.GroundingMetadata{
		SearchEntryPoint: &genai.SearchEntryPoint{RenderedContent: "test"},
	}
	testCitation := &genai.CitationMetadata{
		Citations: []*genai.Citation{{StartIndex: 0, EndIndex: 10}},
	}
	testLogprobs := &genai.LogprobsResult{
		ChosenCandidates: []*genai.LogprobsResultCandidate{{Token: "hello"}},
	}

	tests := []struct {
		name string
		res  *genai.GenerateContentResponse
		want *model.LLMResponse
	}{
		{
			name: "standard candidate with content and full metadata",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: testUsage,
				Candidates: []*genai.Candidate{
					{
						Content:           genai.NewContentFromText("Hello world", genai.RoleModel),
						GroundingMetadata: testGrounding,
						FinishReason:      genai.FinishReasonStop,
						CitationMetadata:  testCitation,
						AvgLogprobs:       0.95,
						LogprobsResult:    testLogprobs,
					},
				},
			},
			want: &model.LLMResponse{
				Content:           genai.NewContentFromText("Hello world", genai.RoleModel),
				GroundingMetadata: testGrounding,
				FinishReason:      genai.FinishReasonStop,
				CitationMetadata:  testCitation,
				AvgLogprobs:       0.95,
				LogprobsResult:    testLogprobs,
				UsageMetadata:     testUsage,
				ModelVersion:      "gemini-2.5-flash",
			},
		},
		{
			name: "clean stop with empty content",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: testUsage,
				Candidates: []*genai.Candidate{
					{
						Content:      &genai.Content{Parts: []*genai.Part{}, Role: genai.RoleModel},
						FinishReason: genai.FinishReasonStop,
					},
				},
			},
			want: &model.LLMResponse{
				Content:       &genai.Content{Parts: []*genai.Part{}, Role: genai.RoleModel},
				FinishReason:  genai.FinishReasonStop,
				UsageMetadata: testUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "blocked or filtered candidate with finish message",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: testUsage,
				Candidates: []*genai.Candidate{
					{
						Content:           &genai.Content{Parts: []*genai.Part{}, Role: genai.RoleModel},
						FinishReason:      genai.FinishReasonSafety,
						FinishMessage:     "Content blocked by safety filter",
						GroundingMetadata: testGrounding,
						CitationMetadata:  testCitation,
						AvgLogprobs:       0.1,
						LogprobsResult:    testLogprobs,
					},
				},
			},
			want: &model.LLMResponse{
				ErrorCode:         string(genai.FinishReasonSafety),
				ErrorMessage:      "Content blocked by safety filter",
				GroundingMetadata: testGrounding,
				FinishReason:      genai.FinishReasonSafety,
				CitationMetadata:  testCitation,
				AvgLogprobs:       0.1,
				LogprobsResult:    testLogprobs,
				UsageMetadata:     testUsage,
				ModelVersion:      "gemini-2.5-flash",
			},
		},
		{
			name: "prompt feedback block without candidates",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: testUsage,
				PromptFeedback: &genai.GenerateContentResponsePromptFeedback{
					BlockReason:        genai.BlockedReasonSafety,
					BlockReasonMessage: "Prompt violates safety guidelines",
				},
			},
			want: &model.LLMResponse{
				ErrorCode:     string(genai.BlockedReasonSafety),
				ErrorMessage:  "Prompt violates safety guidelines",
				UsageMetadata: testUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "empty stream fallback with zero candidates and nil prompt feedback",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-3.1-flash-lite",
				UsageMetadata: testUsage,
			},
			want: &model.LLMResponse{
				Content:       &genai.Content{Parts: []*genai.Part{}, Role: "model"},
				UsageMetadata: testUsage,
				ModelVersion:  "gemini-3.1-flash-lite",
			},
		},
		{
			name: "multi-candidate evaluates first candidate deterministically",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: testUsage,
				Candidates: []*genai.Candidate{
					{
						Content:      genai.NewContentFromText("Candidate 1", genai.RoleModel),
						FinishReason: genai.FinishReasonStop,
					},
					{
						Content:      genai.NewContentFromText("Candidate 2", genai.RoleModel),
						FinishReason: genai.FinishReasonStop,
					},
				},
			},
			want: &model.LLMResponse{
				Content:       genai.NewContentFromText("Candidate 1", genai.RoleModel),
				FinishReason:  genai.FinishReasonStop,
				UsageMetadata: testUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Genai2LLMResponse(tt.res)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Genai2LLMResponse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
