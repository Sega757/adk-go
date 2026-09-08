package openaimodel

import (
	"testing"

	"github.com/openai/openai-go/v3/responses"
	"google.golang.org/genai"
)

func BenchmarkConvertContents(b *testing.B) {
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: "Hello, model!"},
			},
		},
		{
			Role: "model",
			Parts: []*genai.Part{
				{Text: "Hello! How can I help you?"},
			},
		},
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: "Tell me a joke."},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = convertContents(contents)
	}
}

func BenchmarkConvertOutputItems(b *testing.B) {
	items := []responses.ResponseOutputItemUnion{
		{
			Type: "message",
			Content: []responses.ResponseOutputMessageContentUnion{
				{
					Type: "output_text",
					Text: "Here is a response from the OpenAI model.",
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = convertOutputItems(items)
	}
}
