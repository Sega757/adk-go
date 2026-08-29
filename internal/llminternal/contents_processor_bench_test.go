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

package llminternal

import (
	"fmt"
	"testing"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

func generateBenchmarkEvents(chunksPerStream int) []*session.Event {
	events := make([]*session.Event, 0, chunksPerStream*2)

	// Stream 1: User input audio transcription stream
	for i := 0; i < chunksPerStream; i++ {
		events = append(events, &session.Event{
			LLMResponse: model.LLMResponse{
				InputTranscription: &genai.Transcription{
					Text: fmt.Sprintf("streaming user utterance fragment %d ", i),
				},
			},
		})
	}

	// Stream 2: Model output audio transcription stream
	for i := 0; i < chunksPerStream; i++ {
		events = append(events, &session.Event{
			LLMResponse: model.LLMResponse{
				OutputTranscription: &genai.Transcription{
					Text: fmt.Sprintf("synthesized response token chunk %d ", i),
				},
			},
		})
	}

	return events
}

func BenchmarkContentsProcessor_Accumulation(b *testing.B) {
	benchmarks := []struct {
		name       string
		chunkCount int
	}{
		{name: "Chunks_10", chunkCount: 10},
		{name: "Chunks_100", chunkCount: 100},
		{name: "Chunks_500", chunkCount: 500},
	}

	for _, bm := range benchmarks {
		events := generateBenchmarkEvents(bm.chunkCount)

		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				contents, err := buildContentsDefault("testAgent", "", "", events, false, nil)
				if err != nil {
					b.Fatalf("buildContentsDefault failed: %v", err)
				}
				if len(contents) != 2 {
					b.Fatalf("expected 2 aggregated content items, got %d", len(contents))
				}
			}
		})
	}
}
