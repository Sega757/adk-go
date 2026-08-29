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

package conformance_test

import (
	"context"
	"iter"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/internal/configurable"
	"google.golang.org/adk/v2/internal/configurable/conformance"
	"google.golang.org/adk/v2/session"
)

type mockSessionState struct {
	data map[string]any
}

func newMockSessionState() *mockSessionState {
	return &mockSessionState{data: make(map[string]any)}
}

func (s *mockSessionState) Get(key string) (any, error) {
	val, ok := s.data[key]
	if !ok {
		return nil, session.ErrStateKeyNotExist
	}
	return val, nil
}

func (s *mockSessionState) Set(key string, val any) error {
	s.data[key] = val
	return nil
}

func (s *mockSessionState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

type mockSession struct {
	state *mockSessionState
}

func newMockSession() *mockSession {
	return &mockSession{state: newMockSessionState()}
}

func (s *mockSession) ID() string                { return "test-sess-id" }
func (s *mockSession) AppName() string           { return "test-app" }
func (s *mockSession) UserID() string            { return "test-user" }
func (s *mockSession) State() session.State      { return s.state }
func (s *mockSession) Events() session.Events    { return nil }
func (s *mockSession) LastUpdateTime() time.Time { return time.Now() }

type mockAgentContext struct {
	agent.StrictContextMock
	sess session.Session
}

func newMockAgentContext() *mockAgentContext {
	return &mockAgentContext{
		StrictContextMock: agent.NewStrictContextMock(context.Background()),
		sess:              newMockSession(),
	}
}

func (m *mockAgentContext) State() session.State {
	return m.sess.State()
}

func (m *mockAgentContext) Session() session.Session {
	return m.sess
}

func (m *mockAgentContext) InvocationID() string {
	return "inv-123"
}

var expectedCallbackKeys = []string{
	"callback_agent_001.callbacks.before_agent_callback1",
	"callback_agent_001.callbacks.before_agent_callback2",
	"callback_agent_002.callbacks.shortcut_agent_execution",
	"callback_agent_003.callbacks.after_agent_callback1",
	"callback_agent_003.callbacks.after_agent_callback2",
}

func TestRegisterCallbacks(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		err := conformance.RegisterCallbacks()
		if err != nil {
			t.Fatalf("RegisterCallbacks() unexpected error: %v", err)
		}

		ctx := context.Background()
		for _, key := range expectedCallbackKeys {
			cb, err := configurable.ResolveCallbackReference(ctx, key)
			if err != nil {
				t.Errorf("ResolveCallbackReference failed for key %q: %v", key, err)
			}
			if cb == nil {
				t.Errorf("ResolveCallbackReference returned nil callback for key %q", key)
			}
		}
	})

	t.Run("Idempotency_DuplicateRegistrationError", func(t *testing.T) {
		// RegisterCallbacks has already been called in HappyPath and will return error on re-registration
		err := conformance.RegisterCallbacks()
		if err == nil {
			t.Errorf("RegisterCallbacks() expected duplicate registration error, got nil")
		}
	})

	t.Run("Execution_BeforeAgentCallbacks", func(t *testing.T) {
		ctx := context.Background()
		cb1Raw, err := configurable.ResolveCallbackReference(ctx, "callback_agent_001.callbacks.before_agent_callback1")
		if err != nil {
			t.Fatalf("failed to resolve before_agent_callback1: %v", err)
		}
		cb1, ok := cb1Raw.(agent.BeforeAgentCallback)
		if !ok {
			t.Fatalf("before_agent_callback1 has unexpected type %T", cb1Raw)
		}

		cb2Raw, err := configurable.ResolveCallbackReference(ctx, "callback_agent_001.callbacks.before_agent_callback2")
		if err != nil {
			t.Fatalf("failed to resolve before_agent_callback2: %v", err)
		}
		cb2, ok := cb2Raw.(agent.BeforeAgentCallback)
		if !ok {
			t.Fatalf("before_agent_callback2 has unexpected type %T", cb2Raw)
		}

		mockCtx := newMockAgentContext()

		// Run cb1
		res1, err := cb1(mockCtx)
		if err != nil {
			t.Fatalf("cb1 returned error: %v", err)
		}
		if res1 != nil {
			t.Errorf("expected nil Content from cb1, got %v", res1)
		}

		v1, err := mockCtx.State().Get("before_agent_callback_state_key")
		if err != nil {
			t.Fatalf("failed to get state key after cb1: %v", err)
		}
		if v1 != "value1" {
			t.Errorf("expected state 'value1', got %v", v1)
		}

		// Run cb2
		res2, err := cb2(mockCtx)
		if err != nil {
			t.Fatalf("cb2 returned error: %v", err)
		}
		if res2 != nil {
			t.Errorf("expected nil Content from cb2, got %v", res2)
		}

		v2, err := mockCtx.State().Get("before_agent_callback_state_key")
		if err != nil {
			t.Fatalf("failed to get state key after cb2: %v", err)
		}
		if v2 != "value1+value2" {
			t.Errorf("expected state 'value1+value2', got %v", v2)
		}
	})

	t.Run("Execution_ShortcutAgentExecution", func(t *testing.T) {
		ctx := context.Background()
		cbRaw, err := configurable.ResolveCallbackReference(ctx, "callback_agent_002.callbacks.shortcut_agent_execution")
		if err != nil {
			t.Fatalf("failed to resolve shortcut_agent_execution: %v", err)
		}
		cb, ok := cbRaw.(agent.BeforeAgentCallback)
		if !ok {
			t.Fatalf("shortcut_agent_execution has unexpected type %T", cbRaw)
		}

		mockCtx := newMockAgentContext()

		// First invocation when key does not exist -> sets state to "True" and returns nil
		res1, err := cb(mockCtx)
		if err != nil {
			t.Fatalf("shortcut_agent_execution first call error: %v", err)
		}
		if res1 != nil {
			t.Errorf("expected nil Content on first call, got %v", res1)
		}

		v, err := mockCtx.State().Get("conversation_limit_reached")
		if err != nil {
			t.Fatalf("failed to get conversation_limit_reached state: %v", err)
		}
		if v != "True" {
			t.Errorf("expected conversation_limit_reached 'True', got %v", v)
		}

		// Second invocation when key is "True" -> returns shortcut response content
		res2, err := cb(mockCtx)
		if err != nil {
			t.Fatalf("shortcut_agent_execution second call error: %v", err)
		}
		if res2 == nil || len(res2.Parts) == 0 {
			t.Fatalf("expected shortcut Content on second call, got %v", res2)
		}
		if res2.Role != "model" {
			t.Errorf("expected role 'model', got %q", res2.Role)
		}
		expectedText := "Sorry, you have reached the limit of the conversation."
		if res2.Parts[0].Text != expectedText {
			t.Errorf("expected text %q, got %q", expectedText, res2.Parts[0].Text)
		}
	})

	t.Run("Execution_AfterAgentCallbacks", func(t *testing.T) {
		ctx := context.Background()
		cb1Raw, err := configurable.ResolveCallbackReference(ctx, "callback_agent_003.callbacks.after_agent_callback1")
		if err != nil {
			t.Fatalf("failed to resolve after_agent_callback1: %v", err)
		}
		cb1, ok := cb1Raw.(agent.AfterAgentCallback)
		if !ok {
			t.Fatalf("after_agent_callback1 has unexpected type %T", cb1Raw)
		}

		cb2Raw, err := configurable.ResolveCallbackReference(ctx, "callback_agent_003.callbacks.after_agent_callback2")
		if err != nil {
			t.Fatalf("failed to resolve after_agent_callback2: %v", err)
		}
		cb2, ok := cb2Raw.(agent.AfterAgentCallback)
		if !ok {
			t.Fatalf("after_agent_callback2 has unexpected type %T", cb2Raw)
		}

		mockCtx := newMockAgentContext()

		// Run cb1
		res1, err := cb1(mockCtx)
		if err != nil {
			t.Fatalf("after_agent_callback1 returned error: %v", err)
		}
		if res1 != nil {
			t.Errorf("expected nil Content from cb1, got %v", res1)
		}

		v1, err := mockCtx.State().Get("after_agent_callback_state_key")
		if err != nil {
			t.Fatalf("failed to get state key after cb1: %v", err)
		}
		if v1 != "value1" {
			t.Errorf("expected state 'value1', got %v", v1)
		}

		// Run cb2
		res2, err := cb2(mockCtx)
		if err != nil {
			t.Fatalf("after_agent_callback2 returned error: %v", err)
		}
		if res2 != nil {
			t.Errorf("expected nil Content from cb2, got %v", res2)
		}

		v2, err := mockCtx.State().Get("after_agent_callback_state_key")
		if err != nil {
			t.Fatalf("failed to get state key after cb2: %v", err)
		}
		if v2 != "value1+value2" {
			t.Errorf("expected state 'value1+value2', got %v", v2)
		}
	})

	t.Run("Execution_ErrorCasesInCallbacks", func(t *testing.T) {
		ctx := context.Background()
		cb2Raw, err := configurable.ResolveCallbackReference(ctx, "callback_agent_001.callbacks.before_agent_callback2")
		if err != nil {
			t.Fatalf("failed to resolve before_agent_callback2: %v", err)
		}
		cb2 := cb2Raw.(agent.BeforeAgentCallback)

		mockCtx := newMockAgentContext()
		// Set non-string value in state to trigger type assertion error in callback2
		_ = mockCtx.State().Set("before_agent_callback_state_key", 12345)

		_, err = cb2(mockCtx)
		if err == nil {
			t.Errorf("expected error when state value is non-string, got nil")
		}

		afterCb2Raw, err := configurable.ResolveCallbackReference(ctx, "callback_agent_003.callbacks.after_agent_callback2")
		if err != nil {
			t.Fatalf("failed to resolve after_agent_callback2: %v", err)
		}
		afterCb2 := afterCb2Raw.(agent.AfterAgentCallback)

		_ = mockCtx.State().Set("after_agent_callback_state_key", true)
		_, err = afterCb2(mockCtx)
		if err == nil {
			t.Errorf("expected error when state value is non-string in after_agent_callback2, got nil")
		}
	})
}
