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

package runner

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

func BenchmarkFindActiveTaskIsolationScope(b *testing.B) {
	svc := session.InMemoryService()
	ctx := context.Background()
	sessResp, _ := svc.Create(ctx, &session.CreateRequest{
		AppName:   "bench_app",
		UserID:    "u",
		SessionID: "s",
	})
	sess := sessResp.Session

	// Populate session history with 50 events including tool calls and responses.
	for i := 0; i < 50; i++ {
		ev := session.NewEvent(ctx, fmt.Sprintf("inv-%d", i))
		ev.Author = "task_agent"
		ev.IsolationScope = "task-scope-1"
		ev.LLMResponse = model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{FunctionCall: &genai.FunctionCall{Name: "some_tool", ID: fmt.Sprintf("fc-%d", i)}},
				},
			},
		}
		_ = svc.AppendEvent(ctx, sess, ev)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = findActiveTaskIsolationScope(sess)
	}
}

func BenchmarkHandleUserFunctionCallResponse(b *testing.B) {
	svc := session.InMemoryService()
	ctx := context.Background()
	sessResp, _ := svc.Create(ctx, &session.CreateRequest{
		AppName:   "bench_app",
		UserID:    "u",
		SessionID: "s",
	})
	sess := sessResp.Session

	// Populate session history with 50 events.
	for i := 0; i < 50; i++ {
		ev := session.NewEvent(ctx, fmt.Sprintf("inv-%d", i))
		ev.Author = fmt.Sprintf("subagent_%d", i%5)
		ev.LLMResponse = model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{FunctionCall: &genai.FunctionCall{Name: "tool", ID: fmt.Sprintf("call-%d", i)}},
				},
			},
		}
		_ = svc.AppendEvent(ctx, sess, ev)
	}

	msg := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{FunctionResponse: &genai.FunctionResponse{Name: "tool", ID: "call-0"}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	events := sess.Events()
	for i := 0; i < b.N; i++ {
		_ = handleUserFunctionCallResponse(events, msg)
	}
}

func BenchmarkFindEventByFunctionCallID(b *testing.B) {
	svc := session.InMemoryService()
	ctx := context.Background()
	sessResp, _ := svc.Create(ctx, &session.CreateRequest{
		AppName:   "bench_app",
		UserID:    "u",
		SessionID: "s",
	})
	sess := sessResp.Session

	// Populate session history with 50 events.
	for i := 0; i < 50; i++ {
		ev := session.NewEvent(ctx, fmt.Sprintf("inv-%d", i))
		ev.Author = "model"
		ev.LLMResponse = model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{FunctionCall: &genai.FunctionCall{Name: "tool", ID: fmt.Sprintf("call-%d", i)}},
				},
			},
		}
		_ = svc.AppendEvent(ctx, sess, ev)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = findEventByFunctionCallID(sess, "call-0")
	}
}

func BenchmarkOpenLongRunningCallIDs(b *testing.B) {
	svc := session.InMemoryService()
	ctx := context.Background()
	sessResp, _ := svc.Create(ctx, &session.CreateRequest{
		AppName:   "bench_app",
		UserID:    "u",
		SessionID: "s",
	})
	sess := sessResp.Session

	// Populate session history with 50 events.
	for i := 0; i < 50; i++ {
		ev := session.NewEvent(ctx, fmt.Sprintf("inv-%d", i))
		ev.Author = "model"
		if i%5 == 0 {
			ev.LongRunningToolIDs = []string{fmt.Sprintf("lrt-%d", i)}
		}
		ev.LLMResponse = model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{Text: "some response text"},
				},
			},
		}
		_ = svc.AppendEvent(ctx, sess, ev)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = openLongRunningCallIDs(sess)
	}
}
