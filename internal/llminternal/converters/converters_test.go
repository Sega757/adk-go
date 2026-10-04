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

	sampleUsage := &genai.GenerateContentResponseUsageMetadata{
		CandidatesTokenCount: 10,
		PromptTokenCount:     5,
		TotalTokenCount:      15,
	}
	sampleGrounding := &genai.GroundingMetadata{
		SearchEntryPoint: &genai.SearchEntryPoint{RenderedContent: "search rendering"},
	}
	sampleCitation := &genai.CitationMetadata{
		Citations: []*genai.Citation{
			{StartIndex: 0, EndIndex: 10, URI: "https://example.com"},
		},
	}
	sampleLogprobs := &genai.LogprobsResult{
		ChosenCandidates: []*genai.LogprobsResultCandidate{{Token: "hello", LogProbability: -0.1}},
	}
	sampleContent := &genai.Content{
		Role:  "model",
		Parts: []*genai.Part{genai.NewPartFromText("Hello world")},
	}

	tests := []struct {
		name string
		res  *genai.GenerateContentResponse
		want *model.LLMResponse
	}{
		{
			name: "CandidateWithContentAndMetadata",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content:           sampleContent,
						FinishReason:      genai.FinishReasonStop,
						GroundingMetadata: sampleGrounding,
						CitationMetadata:  sampleCitation,
						AvgLogprobs:       -0.05,
						LogprobsResult:    sampleLogprobs,
					},
				},
			},
			want: &model.LLMResponse{
				Content:           sampleContent,
				GroundingMetadata: sampleGrounding,
				FinishReason:      genai.FinishReasonStop,
				CitationMetadata:  sampleCitation,
				AvgLogprobs:       -0.05,
				LogprobsResult:    sampleLogprobs,
				UsageMetadata:     sampleUsage,
				ModelVersion:      "gemini-2.5-flash",
			},
		},
		{
			name: "CandidateWithEmptyContentAndFinishReasonStop",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content:      &genai.Content{Parts: []*genai.Part{}, Role: "model"},
						FinishReason: genai.FinishReasonStop,
					},
				},
			},
			want: &model.LLMResponse{
				Content:       &genai.Content{Parts: []*genai.Part{}, Role: "model"},
				FinishReason:  genai.FinishReasonStop,
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "CandidateWithNonStopFinishReasonAndEmptyParts",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content:           &genai.Content{Parts: []*genai.Part{}, Role: "model"},
						FinishReason:      genai.FinishReasonSafety,
						FinishMessage:     "Blocked by safety filter",
						GroundingMetadata: sampleGrounding,
						CitationMetadata:  sampleCitation,
						AvgLogprobs:       -0.2,
						LogprobsResult:    sampleLogprobs,
					},
				},
			},
			want: &model.LLMResponse{
				ErrorCode:         "SAFETY",
				ErrorMessage:      "Blocked by safety filter",
				GroundingMetadata: sampleGrounding,
				FinishReason:      genai.FinishReasonSafety,
				CitationMetadata:  sampleCitation,
				AvgLogprobs:       -0.2,
				LogprobsResult:    sampleLogprobs,
				UsageMetadata:     sampleUsage,
				ModelVersion:      "gemini-2.5-flash",
			},
		},
		{
			name: "CandidateWithNonStopFinishReasonButHasContentParts",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content:      sampleContent,
						FinishReason: genai.FinishReasonMaxTokens,
					},
				},
			},
			want: &model.LLMResponse{
				Content:       sampleContent,
				FinishReason:  genai.FinishReasonMaxTokens,
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "CandidateWithNilContentAndNonStopFinishReason",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content:       nil,
						FinishReason:  genai.FinishReasonRecitation,
						FinishMessage: "Recitation detected",
					},
				},
			},
			want: &model.LLMResponse{
				ErrorCode:     "RECITATION",
				ErrorMessage:  "Recitation detected",
				FinishReason:  genai.FinishReasonRecitation,
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "MultipleCandidatesEvaluatesFirstCandidate",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				Candidates: []*genai.Candidate{
					{
						Content:      sampleContent,
						FinishReason: genai.FinishReasonStop,
					},
					{
						Content: &genai.Content{
							Role:  "model",
							Parts: []*genai.Part{genai.NewPartFromText("Second candidate")},
						},
						FinishReason: genai.FinishReasonStop,
					},
				},
			},
			want: &model.LLMResponse{
				Content:       sampleContent,
				FinishReason:  genai.FinishReasonStop,
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "PromptFeedbackBlockedResponse",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-2.5-flash",
				UsageMetadata: sampleUsage,
				PromptFeedback: &genai.GenerateContentResponsePromptFeedback{
					BlockReason:        genai.BlockedReasonSafety,
					BlockReasonMessage: "Prompt blocked due to safety settings",
				},
			},
			want: &model.LLMResponse{
				ErrorCode:     "SAFETY",
				ErrorMessage:  "Prompt blocked due to safety settings",
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-2.5-flash",
			},
		},
		{
			name: "EmptyStreamChunkFallback",
			res: &genai.GenerateContentResponse{
				ModelVersion:  "gemini-3.1-flash-lite",
				UsageMetadata: sampleUsage,
			},
			want: &model.LLMResponse{
				Content:       &genai.Content{Parts: []*genai.Part{}, Role: "model"},
				UsageMetadata: sampleUsage,
				ModelVersion:  "gemini-3.1-flash-lite",
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
