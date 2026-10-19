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

package workflowagent

import (
	"context"
	"iter"
	"testing"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

type BenchMockSession struct {
	session.Session
}

func (m BenchMockSession) ID() string {
	return "test-session-id"
}

func (m BenchMockSession) State() session.State {
	return nil
}

func (m BenchMockSession) Events() session.Events {
	return benchEmptySessionEvents{}
}

type benchEmptySessionEvents struct{}

func (benchEmptySessionEvents) At(i int) *session.Event { return nil }
func (benchEmptySessionEvents) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {}
}
func (benchEmptySessionEvents) Len() int { return 0 }

func createTestWorkflowAgent(b *testing.B) *workflowAgent {
	b.Helper()
	nodeA := workflow.NewFunctionNode("dummy", func(ctx agent.Context, input any) (string, error) {
		return "ok", nil
	}, workflow.NodeConfig{})
	w, err := workflow.New("test_workflow", workflow.Chain(workflow.Start, nodeA))
	if err != nil {
		b.Fatalf("failed to create workflow: %v", err)
	}
	return &workflowAgent{workflow: w}
}

func BenchmarkDetectResume_TextTurn(b *testing.B) {
	wa := createTestWorkflowAgent(b)

	mockCtx := &MockInvocationContext{
		Context: context.Background(),
		sess:    BenchMockSession{},
		userContent: &genai.Content{
			Parts: []*genai.Part{{Text: "hello"}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = wa.detectResume(mockCtx)
	}
}

func BenchmarkDetectResume_NonWorkflowFunctionResponse(b *testing.B) {
	wa := createTestWorkflowAgent(b)

	mockCtx := &MockInvocationContext{
		Context: context.Background(),
		sess:    BenchMockSession{},
		userContent: &genai.Content{
			Parts: []*genai.Part{
				{
					FunctionResponse: &genai.FunctionResponse{
						Name:     "get_weather",
						ID:       "call_123",
						Response: map[string]any{"temp": 72},
					},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = wa.detectResume(mockCtx)
	}
}

func BenchmarkDetectResume_WorkflowInputFunctionResponse(b *testing.B) {
	wa := createTestWorkflowAgent(b)

	mockCtx := &MockInvocationContext{
		Context: context.Background(),
		sess:    BenchMockSession{},
		userContent: &genai.Content{
			Parts: []*genai.Part{
				{
					FunctionResponse: &genai.FunctionResponse{
						Name:     workflow.WorkflowInputFunctionCallName,
						ID:       "interrupt_123",
						Response: map[string]any{"response": "user input"},
					},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = wa.detectResume(mockCtx)
	}
}
