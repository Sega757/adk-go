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

package llmagent

import (
	"context"
	"testing"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/internal/workflowinternal"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
)

func BenchmarkFindUnresolvedTaskDelegations_NoDelegations(b *testing.B) {
	ctx := context.Background()
	svc := session.InMemoryService()
	createResp, err := svc.Create(ctx, &session.CreateRequest{
		AppName: "app", UserID: "user", SessionID: "sess-1",
	})
	if err != nil {
		b.Fatalf("session.Create: %v", err)
	}
	sess := createResp.Session

	dummyAgent, err := agent.New(agent.Config{Name: "sub_agent"})
	if err != nil {
		b.Fatalf("agent.New: %v", err)
	}
	taskTool, err := workflowinternal.NewTaskAgentTool(dummyAgent)
	if err != nil {
		b.Fatalf("NewTaskAgentTool: %v", err)
	}
	toolsDict := map[string]tool.Tool{
		taskTool.Name(): taskTool,
	}

	for i := 0; i < 50; i++ {
		ev := session.NewEvent(ctx, "inv-1")
		ev.Author = "user"
		if i%2 == 1 {
			ev.Author = "chat_coordinator"
		}
		ev.LLMResponse = model.LLMResponse{
			Content: &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{{Text: "Hello there, how can I help you today?"}},
			},
		}
		if err := svc.AppendEvent(ctx, sess, ev); err != nil {
			b.Fatalf("AppendEvent: %v", err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = findUnresolvedTaskDelegations(sess, "chat_coordinator", toolsDict)
	}
}

func BenchmarkFindUnresolvedTaskDelegations_ResolvedDelegation(b *testing.B) {
	ctx := context.Background()
	svc := session.InMemoryService()
	createResp, err := svc.Create(ctx, &session.CreateRequest{
		AppName: "app", UserID: "user", SessionID: "sess-2",
	})
	if err != nil {
		b.Fatalf("session.Create: %v", err)
	}
	sess := createResp.Session

	dummyAgent, err := agent.New(agent.Config{Name: "sub_agent"})
	if err != nil {
		b.Fatalf("agent.New: %v", err)
	}
	taskTool, err := workflowinternal.NewTaskAgentTool(dummyAgent)
	if err != nil {
		b.Fatalf("NewTaskAgentTool: %v", err)
	}
	toolsDict := map[string]tool.Tool{
		"sub_agent": taskTool,
	}

	for i := 0; i < 20; i++ {
		ev := session.NewEvent(ctx, "inv-1")
		ev.Author = "user"
		if i%2 == 1 {
			ev.Author = "chat_coordinator"
		}
		ev.LLMResponse = model.LLMResponse{
			Content: &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{{Text: "Chat message"}},
			},
		}
		if err := svc.AppendEvent(ctx, sess, ev); err != nil {
			b.Fatalf("AppendEvent: %v", err)
		}
	}

	// Add resolved task delegation
	fcEv := session.NewEvent(ctx, "inv-2")
	fcEv.Author = "chat_coordinator"
	fcEv.LLMResponse = model.LLMResponse{
		Content: &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{{
				FunctionCall: &genai.FunctionCall{
					ID:   "call-123",
					Name: "sub_agent",
				},
			}},
		},
	}
	if err := svc.AppendEvent(ctx, sess, fcEv); err != nil {
		b.Fatalf("AppendEvent: %v", err)
	}

	frEv := session.NewEvent(ctx, "inv-2")
	frEv.Author = "user"
	frEv.LLMResponse = model.LLMResponse{
		Content: &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					ID:   "call-123",
					Name: "sub_agent",
				},
			}},
		},
	}
	if err := svc.AppendEvent(ctx, sess, frEv); err != nil {
		b.Fatalf("AppendEvent: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = findUnresolvedTaskDelegations(sess, "chat_coordinator", toolsDict)
	}
}
