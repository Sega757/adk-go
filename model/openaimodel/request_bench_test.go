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
