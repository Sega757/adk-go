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
