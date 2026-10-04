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

package llminternal

import (
	"iter"
	"testing"

	"google.golang.org/adk/v2/agent"
	icontext "google.golang.org/adk/v2/internal/context"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

func TestRequestProcessors(t *testing.T) {
	t.Parallel()

	processors := []struct {
		name string
		fn   func(agent.InvocationContext, *model.LLMRequest, *Flow) iter.Seq2[*session.Event, error]
	}{
		{name: "nlPlanningRequestProcessor", fn: nlPlanningRequestProcessor},
		{name: "codeExecutionRequestProcessor", fn: codeExecutionRequestProcessor},
		{name: "authPreprocessor", fn: authPreprocessor},
	}

	for _, tc := range processors {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("non-nil parameters", func(t *testing.T) {
				t.Parallel()

				mockAgent, err := agent.New(agent.Config{Name: "test_agent"})
				if err != nil {
					t.Fatalf("failed to create agent: %v", err)
				}
				ctx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{
					Agent: mockAgent,
				})
				req := &model.LLMRequest{}
				flow := &Flow{}

				seq := tc.fn(ctx, req, flow)
				if seq == nil {
					t.Fatal("expected non-nil iter.Seq2")
				}

				count := 0
				for event, err := range seq {
					count++
					t.Errorf("expected no yield, got event=%v, err=%v", event, err)
				}
				if count != 0 {
					t.Fatalf("expected 0 iterations, got %d", count)
				}
			})

			t.Run("nil parameters", func(t *testing.T) {
				t.Parallel()

				seq := tc.fn(nil, nil, nil)
				if seq == nil {
					t.Fatal("expected non-nil iter.Seq2")
				}

				count := 0
				for event, err := range seq {
					count++
					t.Errorf("expected no yield, got event=%v, err=%v", event, err)
				}
				if count != 0 {
					t.Fatalf("expected 0 iterations, got %d", count)
				}
			})
		})
	}
}

func TestResponseProcessors(t *testing.T) {
	t.Parallel()

	processors := []struct {
		name string
		fn   func(agent.InvocationContext, *model.LLMRequest, *model.LLMResponse) error
	}{
		{name: "nlPlanningResponseProcessor", fn: nlPlanningResponseProcessor},
		{name: "codeExecutionResponseProcessor", fn: codeExecutionResponseProcessor},
	}

	for _, tc := range processors {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("non-nil parameters", func(t *testing.T) {
				t.Parallel()

				mockAgent, err := agent.New(agent.Config{Name: "test_agent"})
				if err != nil {
					t.Fatalf("failed to create agent: %v", err)
				}
				ctx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{
					Agent: mockAgent,
				})
				req := &model.LLMRequest{}
				resp := &model.LLMResponse{}

				if err := tc.fn(ctx, req, resp); err != nil {
					t.Errorf("expected nil error, got: %v", err)
				}
			})

			t.Run("nil parameters", func(t *testing.T) {
				t.Parallel()

				if err := tc.fn(nil, nil, nil); err != nil {
					t.Errorf("expected nil error, got: %v", err)
				}
			})
		})
	}
}
