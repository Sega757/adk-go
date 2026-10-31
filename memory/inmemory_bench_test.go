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

package memory

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"testing"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

type benchSession struct {
	appName, userID, sessionID string
	events                     []*session.Event
}

func (s *benchSession) ID() string                        { return s.sessionID }
func (s *benchSession) AppName() string                   { return s.appName }
func (s *benchSession) UserID() string                    { return s.userID }
func (s *benchSession) Events() session.Events            { return s }
func (s *benchSession) All() iter.Seq[*session.Event]     { return slices.Values(s.events) }
func (s *benchSession) Len() int                          { return len(s.events) }
func (s *benchSession) At(i int) *session.Event           { return s.events[i] }
func (s *benchSession) State() session.State              { panic("not implemented") }
func (s *benchSession) LastUpdateTime() time.Time         { panic("not implemented") }

func makeBenchSession(eventsCount int) session.Session {
	events := make([]*session.Event, eventsCount)
	for i := 0; i < eventsCount; i++ {
		var content *genai.Content
		if i%3 == 0 {
			// Non-text event (e.g. tool call / empty text)
			content = &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{FunctionCall: &genai.FunctionCall{Name: "search"}},
				},
			}
		} else {
			content = genai.NewContentFromText(
				fmt.Sprintf("The quick brown fox jumps over event %d with some additional text words to index", i),
				genai.RoleModel,
			)
		}
		events[i] = &session.Event{
			ID:     fmt.Sprintf("event-%d", i),
			Author: "model",
			LLMResponse: model.LLMResponse{
				Content: content,
			},
			Timestamp: time.Now(),
		}
	}
	return &benchSession{
		appName:   "bench_app",
		userID:    "bench_user",
		sessionID: "bench_sess",
		events:    events,
	}
}

func BenchmarkAddSessionToMemory(b *testing.B) {
	ctx := context.Background()
	sess := makeBenchSession(30)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		svc := InMemoryService()
		if err := svc.AddSessionToMemory(ctx, sess); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSearchMemory(b *testing.B) {
	ctx := context.Background()
	svc := InMemoryService()
	sess := makeBenchSession(30)
	if err := svc.AddSessionToMemory(ctx, sess); err != nil {
		b.Fatal(err)
	}

	req := &SearchRequest{
		AppName: "bench_app",
		UserID:  "bench_user",
		Query:   "quick jumps fox",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := svc.SearchMemory(ctx, req)
		if err != nil || len(resp.Memories) == 0 {
			b.Fatalf("search failed: resp=%v, err=%v", resp, err)
		}
	}
}

func BenchmarkSearchMemory_NotFound(b *testing.B) {
	ctx := context.Background()
	svc := InMemoryService()
	sess := makeBenchSession(30)
	if err := svc.AddSessionToMemory(ctx, sess); err != nil {
		b.Fatal(err)
	}

	req := &SearchRequest{
		AppName: "other_app",
		UserID:  "other_user",
		Query:   "quick jumps fox",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = svc.SearchMemory(ctx, req)
	}
}
