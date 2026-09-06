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

package triggers

import (
	"context"
	"fmt"
	"iter"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/server/adkrest/internal/fakes"
	"google.golang.org/adk/v2/session"
)

func TestRetriableRunner_RunAgent_Success(t *testing.T) {
	runCount := 0
	testAgent, err := agent.New(agent.Config{
		Name: "test-agent",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				runCount++
				yield(&session.Event{ID: "event-1"}, nil)
			}
		},
	})
	if err != nil {
		t.Fatalf("agent.New failed: %v", err)
	}

	sessionService := &fakes.FakeSessionService{Sessions: make(map[fakes.SessionKey]fakes.TestSession)}
	agentLoader := agent.NewSingleLoader(testAgent)

	r := &RetriableRunner{
		sessionService: sessionService,
		agentLoader:    agentLoader,
		triggerConfig: TriggerConfig{
			MaxRetries: 3,
			BaseDelay:  1 * time.Millisecond,
			MaxDelay:   5 * time.Millisecond,
		},
	}

	events, err := r.RunAgent(context.Background(), "test-agent", "user-1", "hello")
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if runCount != 1 {
		t.Errorf("expected 1 run, got %d", runCount)
	}
}

func TestRetriableRunner_RunAgent_ThrottledRetry(t *testing.T) {
	runCount := 0
	throttledErr := fmt.Errorf("429 ResourceExhausted")

	testAgent, err := agent.New(agent.Config{
		Name: "test-agent",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				runCount++
				if runCount < 3 {
					yield(nil, throttledErr)
					return
				}
				yield(&session.Event{ID: "event-ok"}, nil)
			}
		},
	})
	if err != nil {
		t.Fatalf("agent.New failed: %v", err)
	}

	sessionService := &fakes.FakeSessionService{Sessions: make(map[fakes.SessionKey]fakes.TestSession)}
	agentLoader := agent.NewSingleLoader(testAgent)

	r := &RetriableRunner{
		sessionService: sessionService,
		agentLoader:    agentLoader,
		triggerConfig: TriggerConfig{
			MaxRetries: 3,
			BaseDelay:  1 * time.Millisecond,
			MaxDelay:   5 * time.Millisecond,
		},
	}

	events, err := r.RunAgent(context.Background(), "test-agent", "user-1", "hello")
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if runCount != 3 {
		t.Errorf("expected 3 runs, got %d", runCount)
	}
}

func TestRetriableRunner_RunAgent_ContextCancellationDuringBackoff(t *testing.T) {
	runCount := 0
	throttledErr := fmt.Errorf("429 ResourceExhausted")

	testAgent, err := agent.New(agent.Config{
		Name: "test-agent",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				runCount++
				yield(nil, throttledErr)
			}
		},
	})
	if err != nil {
		t.Fatalf("agent.New failed: %v", err)
	}

	sessionService := &fakes.FakeSessionService{Sessions: make(map[fakes.SessionKey]fakes.TestSession)}
	agentLoader := agent.NewSingleLoader(testAgent)

	// Set a long BaseDelay so that blocking time.Sleep would take 500ms
	r := &RetriableRunner{
		sessionService: sessionService,
		agentLoader:    agentLoader,
		triggerConfig: TriggerConfig{
			MaxRetries: 5,
			BaseDelay:  500 * time.Millisecond,
			MaxDelay:   2 * time.Second,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context during backoff wait after first run
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err = r.RunAgent(ctx, "test-agent", "user-1", "hello")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected error on context cancellation, got nil")
	}

	t.Logf("RunAgent returned after %v with error: %v", elapsed, err)
}

func BenchmarkRunAgentWithRetry_ContextCancellation(b *testing.B) {
	throttledErr := fmt.Errorf("429 ResourceExhausted")

	testAgent, err := agent.New(agent.Config{
		Name: "test-agent",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				yield(nil, throttledErr)
			}
		},
	})
	if err != nil {
		b.Fatalf("agent.New failed: %v", err)
	}

	sessionService := &fakes.FakeSessionService{Sessions: make(map[fakes.SessionKey]fakes.TestSession)}
	agentLoader := agent.NewSingleLoader(testAgent)

	r := &RetriableRunner{
		sessionService: sessionService,
		agentLoader:    agentLoader,
		triggerConfig: TriggerConfig{
			MaxRetries: 5,
			BaseDelay:  200 * time.Millisecond,
			MaxDelay:   1 * time.Second,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(1 * time.Millisecond)
			cancel()
		}()
		_, _ = r.RunAgent(ctx, "test-agent", "user-1", "hello")
	}
}
