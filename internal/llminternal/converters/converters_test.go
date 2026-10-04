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

package converters

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genai"

	"google.golang.org/adk/v2/model"
)

func TestGenai2LLMResponse(t *testing.T) {
	sampleUsage := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     10,
		CandidatesTokenCount: 20,
		TotalTokenCount:      30,
	}
	sampleGrounding := &genai.GroundingMetadata{
		WebSearchQueries: []string{"test query"},
	}
	sampleCitation := &genai.CitationMetadata{
		Citations: []*genai.Citation{
			{StartIndex: 0, EndIndex: 5},
		},
	}
	sampleAvgLogprobs := 0.95
	sampleLogprobsResult := &genai.LogprobsResult{
		ChosenCandidates: []*genai.LogprobsResultCandidate{
			{Token: "test", LogProbability: -0.1},
		},
	}

	tests := []struct {
		name     string
		input    *genai.GenerateContentResponse
		expected *model.LLMResponse
	}{
		{
			name: "candidate with content parts and finish reason stop",
			input: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Role:  "model",
							Parts: []*genai.Part{genai.NewPartFromText("hello")},
						},
						FinishReason:      genai.FinishReasonStop,
						GroundingMetadata: sampleGrounding,
						CitationMetadata:  sampleCitation,
						AvgLogprobs:       sampleAvgLogprobs,
						LogprobsResult:    sampleLogprobsResult,
					},
				},
			},
			expected: &model.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{genai.NewPartFromText("hello")},
				},
				FinishReason:      genai.FinishReasonStop,
				GroundingMetadata: sampleGrounding,
				CitationMetadata:  sampleCitation,
				AvgLogprobs:       sampleAvgLogprobs,
				LogprobsResult:    sampleLogprobsResult,
				UsageMetadata:     sampleUsage,
				ModelVersion:      "gemini-2.5-flash",
			},
		},
		{
			name: "candidate with parts but non-stop finish reason",
			input: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Role:  "model",
							Parts: []*genai.Part{genai.NewPartFromText("partial text")},
						},
						FinishReason: genai.FinishReasonMaxTokens,
					},
				},
			},
			expected: &model.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{genai.NewPartFromText("partial text")},
				},
				FinishReason:  genai.FinishReasonMaxTokens,
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "candidate without parts but finish reason stop",
			input: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content:      &genai.Content{Role: "model", Parts: []*genai.Part{}},
						FinishReason: genai.FinishReasonStop,
					},
				},
			},
			expected: &model.LLMResponse{
				Content:       &genai.Content{Role: "model", Parts: []*genai.Part{}},
				FinishReason:  genai.FinishReasonStop,
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "candidate error or blocked finish reason without content parts",
			input: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						FinishReason:      genai.FinishReasonSafety,
						FinishMessage:     "content flagged by safety filters",
						GroundingMetadata: sampleGrounding,
						CitationMetadata:  sampleCitation,
						AvgLogprobs:       sampleAvgLogprobs,
						LogprobsResult:    sampleLogprobsResult,
					},
				},
			},
			expected: &model.LLMResponse{
				ErrorCode:         string(genai.FinishReasonSafety),
				ErrorMessage:      "content flagged by safety filters",
				FinishReason:      genai.FinishReasonSafety,
				GroundingMetadata: sampleGrounding,
				CitationMetadata:  sampleCitation,
				AvgLogprobs:       sampleAvgLogprobs,
				LogprobsResult:    sampleLogprobsResult,
				UsageMetadata:     sampleUsage,
				ModelVersion:      "gemini-2.5-flash",
			},
		},
		{
			name: "prompt feedback block reason",
			input: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				PromptFeedback: &genai.GenerateContentResponsePromptFeedback{
					BlockReason:        genai.BlockedReasonSafety,
					BlockReasonMessage: "prompt blocked due to safety",
				},
			},
			expected: &model.LLMResponse{
				ErrorCode:     string(genai.BlockedReasonSafety),
				ErrorMessage:  "prompt blocked due to safety",
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "empty response fallback with no candidates and no prompt feedback",
			input: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-3.1-flash-lite",
				UsageMetadata: sampleUsage,
			},
			expected: &model.LLMResponse{
				Content:       &genai.Content{Parts: []*genai.Part{}, Role: "model"},
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-3.1-flash-lite",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Genai2LLMResponse(tc.input)
			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("Genai2LLMResponse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
