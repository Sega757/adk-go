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

package main

import (
	"context"
	"iter"

	"google.golang.org/genai"
	"google.golang.org/adk/model"
)

type MockModel struct{}

func (m *MockModel) Name() string {
	return "mock-model"
}

func (m *MockModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		response := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{
						Text: "Greetings! I am your ADK 2.0 Starter Kit agent. I can help you with time, weather, and even generating ideas!",
					},
				},
			},
			TurnComplete: true,
		}
		yield(response, nil)
	}
}
