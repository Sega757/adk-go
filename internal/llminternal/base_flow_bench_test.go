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
	"testing"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/tool"
)

// Legacy simulator of task_completed sleep behavior.
func simulateTaskCompletedLegacy() {
	isTaskCompleted := true
	if isTaskCompleted {
		time.Sleep(100 * time.Millisecond)
		return
	}
}

// Optimized simulator without time.Sleep.
func simulateTaskCompletedOptimized() {
	isTaskCompleted := true
	if isTaskCompleted {
		return
	}
}

func BenchmarkTaskCompleted_Legacy(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		simulateTaskCompletedLegacy()
	}
}

func BenchmarkTaskCompleted_Optimized(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		simulateTaskCompletedOptimized()
	}
}

func BenchmarkFindLongRunningFunctionCallIDs_NoCalls(b *testing.B) {
	c := &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			{Text: "Hello world text response"},
		},
	}
	tools := map[string]tool.Tool{}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = findLongRunningFunctionCallIDs(c, tools)
	}
}

func BenchmarkFindLongRunningFunctionCallIDs_StandardCall(b *testing.B) {
	c := &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			{
				FunctionCall: &genai.FunctionCall{
					ID:   "call-1",
					Name: "standard_tool",
				},
			},
		},
	}
	tools := map[string]tool.Tool{
		"standard_tool": &mockLongRunningTool{name: "standard_tool", isLongRunning: false},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = findLongRunningFunctionCallIDs(c, tools)
	}
}

func BenchmarkFindLongRunningFunctionCallIDs_LongRunningCall(b *testing.B) {
	c := &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			{
				FunctionCall: &genai.FunctionCall{
					ID:   "call-1",
					Name: "long_tool",
				},
			},
		},
	}
	tools := map[string]tool.Tool{
		"long_tool": &mockLongRunningTool{name: "long_tool", isLongRunning: true},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = findLongRunningFunctionCallIDs(c, tools)
	}
}

func TestTaskCompleted_PartMatching(t *testing.T) {
	parts := []*genai.Part{
		{
			FunctionResponse: &genai.FunctionResponse{
				Name: "task_completed",
			},
		},
	}
	var isTaskCompleted bool
	for _, part := range parts {
		if part.FunctionResponse != nil && part.FunctionResponse.Name == "task_completed" {
			isTaskCompleted = true
			break
		}
	}
	if !isTaskCompleted {
		t.Fatalf("expected task_completed to be identified")
	}
}
