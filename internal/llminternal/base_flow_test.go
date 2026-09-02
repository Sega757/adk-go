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
	"context"
	"errors"
	"google.golang.org/adk/v2/internal/agent/runconfig"
	"google.golang.org/adk/v2/tool/functiontool"
	"iter"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	icontext "google.golang.org/adk/v2/internal/context"
	"google.golang.org/adk/v2/internal/toolinternal"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
)

type mockFunctionTool struct {
	name    string
	runFunc func(agent.Context, map[string]any) (map[string]any, error)
}

func (m *mockFunctionTool) Name() string {
	return m.name
}

func (m *mockFunctionTool) Description() string {
	return "mock tool"
}

func (m *mockFunctionTool) InputSchema() *genai.Schema {
	return nil
}

func (m *mockFunctionTool) OutputSchema() *genai.Schema {
	return nil
}

func (m *mockFunctionTool) IsLongRunning() bool {
	return false
}

func (m *mockFunctionTool) ProcessRequest(ctx agent.Context, req *model.LLMRequest) error {
	return nil
}

func (m *mockFunctionTool) Run(ctx agent.Context, args any) (map[string]any, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, args.(map[string]any))
	}
	return nil, nil
}

func (m *mockFunctionTool) Declaration() *genai.FunctionDeclaration {
	return nil
}

type mockToolset struct {
	name string
}

func (m *mockToolset) Name() string { return m.name }
func (m *mockToolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	return nil, nil
}

type mockRequestProcessorToolset struct {
	name    string
	process func(ctx agent.Context, req *model.LLMRequest) error
}

func (m *mockRequestProcessorToolset) ProcessRequest(ctx agent.Context, req *model.LLMRequest) error {
	if m.process != nil {
		return m.process(ctx, req)
	}
	return nil
}
func (m *mockRequestProcessorToolset) Name() string { return m.name }
func (m *mockRequestProcessorToolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	return nil, nil
}

type testCase struct {
	name                 string
	tool                 toolinternal.FunctionTool
	args                 map[string]any
	beforeToolCallbacks  []BeforeToolCallback
	afterToolCallbacks   []AfterToolCallback
	onToolErrorCallbacks []OnToolErrorCallback
	want                 map[string]any
}

func TestCallTool(t *testing.T) {
	testCases := []testCase{
		{
			name: "tool runs successfully",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					return map[string]any{"result": "success"}, nil
				},
			},
			args: map[string]any{"key": "value"},
			want: map[string]any{"result": "success"},
		},
		{
			name: "tool error",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					return nil, errors.New("tool error")
				},
			},
			args: map[string]any{"key": "value"},
			want: map[string]any{"error": "tool error"},
		},
		{
			name: "before callback returns result",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return map[string]any{"result": "intercepted"}, nil
				},
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return map[string]any{"result": "2nd callback should not be called"}, nil
				},
			},
			want: map[string]any{"result": "intercepted"},
		},
		{
			name: "before callback returns error",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, errors.New("before callback error")
				},
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, errors.New("unexpected error")
				},
			},
			want: map[string]any{"error": "before callback error"},
		},
		{
			name: "after callback modifies result",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					return map[string]any{"result": "original"}, nil
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					return map[string]any{"result": "modified"}, nil
				},
			},
			want: map[string]any{"result": "modified"},
		},
		{
			name: "after callback handles error",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					return nil, errors.New("tool error")
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					if err != nil {
						return map[string]any{"result": "error handled"}, nil
					}
					return nil, nil
				},
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					return map[string]any{"result": "unexpected output"}, nil
				},
			},
			want: map[string]any{"result": "error handled"},
		},
		{
			name: "after callback returns error",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					return map[string]any{"result": "success"}, nil
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					return nil, errors.New("after callback error")
				},
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					return nil, errors.New("unexpected error")
				},
			},
			want: map[string]any{"error": "after callback error"},
		},
		{
			name: "no-op callbacks return func results",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					return map[string]any{"result": "success"}, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, nil
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					return nil, nil
				},
			},
			want: map[string]any{"result": "success"},
		},
		{
			name: "before callback result passed to after callback",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return map[string]any{"result": "from_before"}, nil
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					if val, ok := result["result"]; !ok || val != "from_before" {
						return nil, errors.New("unexpected result in after callback")
					}
					return map[string]any{"result": "from_after"}, nil
				},
			},
			want: map[string]any{"result": "from_after"},
		},
		{
			name: "before callback error passed to after callback",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, errors.New("error_from_before")
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					if err == nil || err.Error() != "error_from_before" {
						return nil, errors.New("unexpected error in after callback")
					}
					return map[string]any{"result": "error_handled_in_after"}, nil
				},
			},
			want: map[string]any{"result": "error_handled_in_after"},
		},
		{
			name: "before callback error passed to on tool error callback",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, errors.New("error_from_before")
				},
			},
			onToolErrorCallbacks: []OnToolErrorCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any, err error) (map[string]any, error) {
					if err == nil || err.Error() != "error_from_before" {
						t.Error("unexpected error in on tool error callback")
						return nil, errors.New("unexpected error in on tool error callback")
					}
					return map[string]any{"result": "error_handled_in_on_tool_error_callback"}, nil
				},
			},
			want: map[string]any{"result": "error_handled_in_on_tool_error_callback"},
		},
		{
			name: "before callback error passed to on tool error callback and after tool called",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, errors.New("error_from_before")
				},
			},
			onToolErrorCallbacks: []OnToolErrorCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any, err error) (map[string]any, error) {
					if err == nil || err.Error() != "error_from_before" {
						t.Error("unexpected error in on tool error callback")
						return nil, errors.New("unexpected error in on tool error callback")
					}
					return map[string]any{"result": "error_handled_in_on_tool_error_callback"}, nil
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					if err != nil {
						return nil, errors.New("unexpected error in after callback")
					}
					return map[string]any{"result": "from_after"}, nil
				},
			},
			want: map[string]any{"result": "from_after"},
		},
		{
			name: "before callback error passed to on tool error callback and passed to after tool called",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, errors.New("error_from_before")
				},
			},
			onToolErrorCallbacks: []OnToolErrorCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any, err error) (map[string]any, error) {
					if err == nil || err.Error() != "error_from_before" {
						t.Error("unexpected error in on tool error callback")
						return nil, errors.New("unexpected error in on tool error callback")
					}
					return nil, errors.New("error_from_on_tool_error")
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					if err == nil || err.Error() != "error_from_on_tool_error" {
						return nil, errors.New("unexpected error in after callback")
					}
					return nil, errors.New("error_from_after_tool")
				},
			},
			want: map[string]any{"error": "error_from_after_tool"},
		},
		{
			name: "before callback error passed to on tool error callback and passed to after tool called and handled",
			tool: &mockFunctionTool{
				name: "testTool",
				runFunc: func(ctx agent.Context, args map[string]any) (map[string]any, error) {
					t.Error("tool should not be called")
					return nil, nil
				},
			},
			beforeToolCallbacks: []BeforeToolCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any) (map[string]any, error) {
					return nil, errors.New("error_from_before")
				},
			},
			onToolErrorCallbacks: []OnToolErrorCallback{
				func(ctx agent.Context, tool tool.Tool, args map[string]any, err error) (map[string]any, error) {
					if err == nil || err.Error() != "error_from_before" {
						t.Error("unexpected error in on tool error callback")
						return nil, errors.New("unexpected error in on tool error callback")
					}
					return nil, errors.New("error_from_on_tool_error")
				},
			},
			afterToolCallbacks: []AfterToolCallback{
				func(ctx agent.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
					if err == nil || err.Error() != "error_from_on_tool_error" {
						return nil, errors.New("unexpected error in after callback")
					}
					return map[string]any{"result": "error_handled_in_on_tool_error_callback"}, nil
				},
			},
			want: map[string]any{"result": "error_handled_in_on_tool_error_callback"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := &Flow{
				BeforeToolCallbacks:  tc.beforeToolCallbacks,
				AfterToolCallbacks:   tc.afterToolCallbacks,
				OnToolErrorCallbacks: tc.onToolErrorCallbacks,
			}
			ctx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{})
			got := f.callTool(agent.NewToolContext(ctx, "", nil, nil), tc.tool, tc.args)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("callTool() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkDeepMergeMap(b *testing.B) {
	src := map[string]any{
		"k1": "v1",
		"k2": "v2",
		"k3": map[string]any{
			"sub1": "sv1",
			"sub2": "sv2",
		},
		"k4": 100,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = deepMergeMap(nil, src)
	}
}

func BenchmarkMergeEventActions(b *testing.B) {
	base := &session.EventActions{
		StateDelta: map[string]any{"a": 1, "b": 2},
	}
	other := &session.EventActions{
		StateDelta: map[string]any{"b": 3, "c": map[string]any{"d": 4}},
		Escalate:   true,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = mergeEventActions(base, other)
	}
}

func TestMergeEventActions(t *testing.T) {
	tests := []struct {
		name  string
		base  *session.EventActions
		other *session.EventActions
		want  *session.EventActions
	}{
		{
			name:  "both nil",
			base:  nil,
			other: nil,
			want:  nil,
		},
		{
			name: "other nil returns base",
			base: &session.EventActions{
				StateDelta: map[string]any{"key1": "value1"},
			},
			other: nil,
			want: &session.EventActions{
				StateDelta: map[string]any{"key1": "value1"},
			},
		},
		{
			name: "base nil returns other",
			base: nil,
			other: &session.EventActions{
				StateDelta: map[string]any{"key1": "value1"},
			},
			want: &session.EventActions{
				StateDelta: map[string]any{"key1": "value1"},
			},
		},
		{
			name: "state delta merged with non-overlapping keys",
			base: &session.EventActions{
				StateDelta: map[string]any{"key1": "value1"},
			},
			other: &session.EventActions{
				StateDelta: map[string]any{"key2": "value2"},
			},
			want: &session.EventActions{
				StateDelta: map[string]any{"key1": "value1", "key2": "value2"},
			},
		},
		{
			name: "state delta merged with overlapping keys - later wins",
			base: &session.EventActions{
				StateDelta: map[string]any{"key1": "original"},
			},
			other: &session.EventActions{
				StateDelta: map[string]any{"key1": "overwritten"},
			},
			want: &session.EventActions{
				StateDelta: map[string]any{"key1": "overwritten"},
			},
		},
		{
			name: "state delta merged with nested map values",
			base: &session.EventActions{
				StateDelta: map[string]any{
					"outer": map[string]any{"key1": "value1", "key2": "value2"},
				},
			},
			other: &session.EventActions{
				StateDelta: map[string]any{
					"outer": map[string]any{"key2": "updated", "key3": "value3"},
				},
			},
			want: &session.EventActions{
				StateDelta: map[string]any{
					"outer": map[string]any{"key1": "value1", "key2": "updated", "key3": "value3"},
				},
			},
		},
		{
			name: "state delta merged with multiple keys from multiple tools",
			base: &session.EventActions{
				StateDelta: map[string]any{"tool1_key": "tool1_value"},
			},
			other: &session.EventActions{
				StateDelta: map[string]any{"tool2_key": "tool2_value", "tool3_key": "tool3_value"},
			},
			want: &session.EventActions{
				StateDelta: map[string]any{
					"tool1_key": "tool1_value",
					"tool2_key": "tool2_value",
					"tool3_key": "tool3_value",
				},
			},
		},
		{
			name: "base has nil state delta, other has values",
			base: &session.EventActions{
				SkipSummarization: true,
			},
			other: &session.EventActions{
				StateDelta: map[string]any{"key1": "value1"},
			},
			want: &session.EventActions{
				SkipSummarization: true,
				StateDelta:        map[string]any{"key1": "value1"},
			},
		},
		{
			name: "skip summarization merging - any true wins",
			base: &session.EventActions{
				SkipSummarization: false,
			},
			other: &session.EventActions{
				SkipSummarization: true,
			},
			want: &session.EventActions{
				SkipSummarization: true,
			},
		},
		{
			name: "escalate merging - any true wins",
			base: &session.EventActions{
				Escalate: false,
			},
			other: &session.EventActions{
				Escalate: true,
			},
			want: &session.EventActions{
				Escalate: true,
			},
		},
		{
			name: "transfer to agent - last wins",
			base: &session.EventActions{
				TransferToAgent: "agent1",
			},
			other: &session.EventActions{
				TransferToAgent: "agent2",
			},
			want: &session.EventActions{
				TransferToAgent: "agent2",
			},
		},
		{
			name: "all fields merged correctly",
			base: &session.EventActions{
				StateDelta:        map[string]any{"key1": "value1"},
				SkipSummarization: false,
				TransferToAgent:   "agent1",
				Escalate:          false,
			},
			other: &session.EventActions{
				StateDelta:        map[string]any{"key2": "value2"},
				SkipSummarization: true,
				TransferToAgent:   "agent2",
				Escalate:          true,
			},
			want: &session.EventActions{
				StateDelta:        map[string]any{"key1": "value1", "key2": "value2"},
				SkipSummarization: true,
				TransferToAgent:   "agent2",
				Escalate:          true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeEventActions(tc.base, tc.other)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mergeEventActions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPreprocess_Toolset(t *testing.T) {
	noOpAgent, err := agent.New(agent.Config{Name: "no-op"})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	tests := []struct {
		name      string
		agent     agent.Agent
		wantModel string
		wantError bool
	}{
		{
			name:      "agent not llminternal.Agent",
			agent:     noOpAgent,
			wantError: false,
		},
		{
			name:      "agent has no toolsets",
			agent:     &mockLLMAgent{s: &State{}},
			wantError: false,
		},
		{
			name: "toolset implements RequestProcessor, error",
			agent: &mockLLMAgent{
				s: &State{
					Toolsets: []tool.Toolset{&mockRequestProcessorToolset{
						name: "toolset",
						process: func(_ agent.Context, _ *model.LLMRequest) error {
							return errors.New("process error")
						},
					}},
				},
			},
			wantError: true,
		},
		{
			name: "toolsets, success",
			agent: &mockLLMAgent{
				s: &State{
					Toolsets: []tool.Toolset{
						&mockToolset{name: "toolset_without_processor"},
						&mockRequestProcessorToolset{
							name: "toolset_with_processor",
							process: func(_ agent.Context, req *model.LLMRequest) error {
								req.Model = "modified-model"
								return nil
							},
						},
					},
				},
			},
			wantError: false,
			wantModel: "modified-model",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Flow{}
			ctx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{Agent: tc.agent})
			req := &model.LLMRequest{}

			events := f.preprocess(ctx, req)

			var gotErr error
			for _, err := range events {
				if err != nil {
					gotErr = err
					break
				}
			}
			if (gotErr != nil) != tc.wantError {
				t.Errorf("preprocess() error = %v, wantError %v", gotErr, tc.wantError)
			}
			if req.Model != tc.wantModel {
				t.Errorf("preprocess() model = %s, wantModel %s", req.Model, tc.wantModel)
			}
		})
	}
}

// fnRespEvent builds a function-response event carrying a single text part,
// mirroring what the parallel-call producer emits for a normal tool.
func fnRespEvent(t *testing.T, text string) *session.Event {
	t.Helper()
	ev := session.NewEvent(t.Context(), "inv")
	ev.LLMResponse = model.LLMResponse{
		Content: &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: text}},
		},
	}
	return ev
}

// TestMergeParallelFunctionResponseEvents_NilEntries guards that nil slots
// (left by long-running/deferred tools that return early) don't panic the
// merge. Pre-fix, a nil events[0] or an all-nil slice dereferenced nil.
func TestMergeParallelFunctionResponseEvents_NilEntries(t *testing.T) {
	t.Run("first entry nil", func(t *testing.T) {
		got, err := mergeParallelFunctionResponseEvents([]*session.Event{nil, fnRespEvent(t, "b")})
		if err != nil {
			t.Fatalf("merge error: %v", err)
		}
		if got == nil || got.LLMResponse.Content == nil {
			t.Fatalf("got nil/empty merged event: %#v", got)
		}
		if n := len(got.LLMResponse.Content.Parts); n != 1 {
			t.Errorf("merged parts = %d, want 1", n)
		}
	})

	t.Run("all entries nil", func(t *testing.T) {
		got, err := mergeParallelFunctionResponseEvents([]*session.Event{nil, nil})
		if err != nil {
			t.Fatalf("merge error: %v", err)
		}
		if got != nil {
			t.Errorf("merged event = %#v, want nil", got)
		}
	})

	t.Run("mixed nil and non-nil", func(t *testing.T) {
		got, err := mergeParallelFunctionResponseEvents([]*session.Event{fnRespEvent(t, "a"), nil, fnRespEvent(t, "c")})
		if err != nil {
			t.Fatalf("merge error: %v", err)
		}
		if got == nil || got.LLMResponse.Content == nil {
			t.Fatalf("got nil/empty merged event: %#v", got)
		}
		if n := len(got.LLMResponse.Content.Parts); n != 2 {
			t.Errorf("merged parts = %d, want 2", n)
		}
	})
}

func TestIsThoughtOnlyTurn(t *testing.T) {
	event := func(partial bool, parts ...*genai.Part) *session.Event {
		var content *genai.Content
		if parts != nil {
			content = &genai.Content{Role: "model", Parts: parts}
		}
		return &session.Event{LLMResponse: model.LLMResponse{Content: content, Partial: partial}}
	}

	tests := []struct {
		name string
		ev   *session.Event
		want bool
	}{
		{"thought_text_only", event(false, &genai.Part{Thought: true, Text: "thinking"}), true},
		{"thought_plus_answer", event(false, &genai.Part{Thought: true, Text: "t"}, &genai.Part{Text: "answer"}), false},
		{"answer_only", event(false, &genai.Part{Text: "answer"}), false},
		{"function_call", event(false, &genai.Part{FunctionCall: &genai.FunctionCall{Name: "f"}}), false},
		{"thought_then_signed_call", event(false, &genai.Part{Thought: true, Text: "t"}, &genai.Part{FunctionCall: &genai.FunctionCall{Name: "f"}, ThoughtSignature: []byte("sig")}), false},
		{"thought_plus_signature_only_part", event(false, &genai.Part{Thought: true, Text: "t"}, &genai.Part{ThoughtSignature: []byte("sig")}), false},
		{"partial_thought", event(true, &genai.Part{Thought: true, Text: "t"}), false},
		{"empty_content", event(false), false},
		{"nil_event", nil, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isThoughtOnlyTurn(tc.ev); got != tc.want {
				t.Errorf("isThoughtOnlyTurn = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExtractTools(t *testing.T) {
	mock1 := &mockFunctionTool{name: "tool1"}
	mock2 := &mockFunctionTool{name: "tool2"}

	tests := []struct {
		name     string
		toolsMap map[string]any
		want     map[string]tool.Tool
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "nil map",
			toolsMap: nil,
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "empty map",
			toolsMap: map[string]any{},
			want:     nil,
			wantErr:  false,
		},
		{
			name: "valid tool map",
			toolsMap: map[string]any{
				"tool1": mock1,
				"tool2": mock2,
			},
			want: map[string]tool.Tool{
				"tool1": mock1,
				"tool2": mock2,
			},
			wantErr: false,
		},
		{
			name: "invalid tool type",
			toolsMap: map[string]any{
				"validTool": mock1,
				"badTool":   "not-a-tool-instance",
			},
			want:    nil,
			wantErr: true,
			errMsg:  "unexpected tool type string for tool badTool",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractTools(tc.toolsMap)
			if (err != nil) != tc.wantErr {
				t.Fatalf("extractTools() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if err.Error() != tc.errMsg {
					t.Errorf("extractTools() error msg = %q, want %q", err.Error(), tc.errMsg)
				}
				return
			}
			if diff := cmp.Diff(tc.want, got, cmp.AllowUnexported(mockFunctionTool{})); diff != "" {
				t.Errorf("extractTools() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type mockLLMWithTools struct {
	name string
}

func (m *mockLLMWithTools) Name() string { return m.name }
func (m *mockLLMWithTools) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(&model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{{Text: "hello"}},
			},
		}, nil)
	}
}

func TestRunOneStep_InvalidToolType(t *testing.T) {
	invalidProcessor := func(ctx agent.InvocationContext, req *model.LLMRequest, f *Flow) iter.Seq2[*session.Event, error] {
		return func(yield func(*session.Event, error) bool) {
			if req.Tools == nil {
				req.Tools = make(map[string]any)
			}
			req.Tools["invalid_tool"] = 12345
		}
	}

	mockAgent, err := agent.New(agent.Config{Name: "test-agent"})
	if err != nil {
		t.Fatalf("agent.New failed: %v", err)
	}

	f := &Flow{
		Model:             &mockLLMWithTools{name: "test-model"},
		RequestProcessors: []func(ctx agent.InvocationContext, req *model.LLMRequest, f *Flow) iter.Seq2[*session.Event, error]{invalidProcessor},
	}

	ctx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{Agent: mockAgent})

	var gotErr error
	for _, err := range f.runOneStep(ctx) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected error for invalid tool type, got nil")
	}
	wantMsg := "unexpected tool type int for tool invalid_tool"
	if gotErr.Error() != wantMsg {
		t.Errorf("error = %q, want %q", gotErr.Error(), wantMsg)
	}
}

func TestHandleFunctionCalls_ToolTimeout(t *testing.T) {
	slowSyncTool, err := functiontool.New(functiontool.Config{
		Name:        "slow_sync",
		Description: "slow sync tool",
	}, func(ctx agent.Context, args map[string]any) (map[string]any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return map[string]any{"result": "ok"}, nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	slowStreamTool, err := functiontool.NewStreaming(functiontool.Config{
		Name:        "slow_stream",
		Description: "slow stream tool",
	}, func(ctx agent.Context, args map[string]any) iter.Seq2[string, error] {
		return func(yield func(string, error) bool) {
			select {
			case <-ctx.Done():
				yield("", ctx.Err())
				return
			case <-time.After(200 * time.Millisecond):
				yield("chunk1", nil)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	toolsDict := map[string]tool.Tool{
		"slow_sync":   slowSyncTool,
		"slow_stream": slowStreamTool,
	}

	t.Run("Configured ToolTimeout on Synchronous Tool", func(t *testing.T) {
		runCfg := &runconfig.RunConfig{ToolTimeout: 30 * time.Millisecond}
		parentCtx := runconfig.ToContext(t.Context(), runCfg)
		invCtx := icontext.NewInvocationContext(parentCtx, icontext.InvocationContextParams{
			InvocationID: "inv_timeout_1",
			Agent:        &mockAgent{name: "agent_1"},
		})

		flow := &Flow{Tools: []tool.Tool{slowSyncTool}}
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{FunctionCall: &genai.FunctionCall{ID: "c1", Name: "slow_sync", Args: map[string]any{}}},
				},
			},
		}

		mergedEvent, err := flow.handleFunctionCalls(invCtx, toolsDict, resp, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mergedEvent == nil || len(mergedEvent.LLMResponse.Content.Parts) != 1 {
			t.Fatalf("invalid merged event: %#v", mergedEvent)
		}

		fr := mergedEvent.LLMResponse.Content.Parts[0].FunctionResponse
		if fr == nil {
			t.Fatal("expected function response")
		}
		errStr, ok := fr.Response["error"].(string)
		if !ok || !strings.Contains(errStr, "tool execution timed out") {
			t.Errorf("expected timeout error in response, got: %v", fr.Response)
		}
	})

	t.Run("Configured ToolTimeout on Streaming Tool", func(t *testing.T) {
		runCfg := &runconfig.RunConfig{ToolTimeout: 30 * time.Millisecond}
		parentCtx := runconfig.ToContext(t.Context(), runCfg)
		invCtx := icontext.NewInvocationContext(parentCtx, icontext.InvocationContextParams{
			InvocationID: "inv_timeout_2",
			Agent:        &mockAgent{name: "agent_2"},
		})

		flow := &Flow{Tools: []tool.Tool{slowStreamTool}}
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{FunctionCall: &genai.FunctionCall{ID: "c2", Name: "slow_stream", Args: map[string]any{}}},
				},
			},
		}

		mergedEvent, err := flow.handleFunctionCalls(invCtx, toolsDict, resp, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mergedEvent == nil || len(mergedEvent.LLMResponse.Content.Parts) != 1 {
			t.Fatalf("invalid merged event: %#v", mergedEvent)
		}

		fr := mergedEvent.LLMResponse.Content.Parts[0].FunctionResponse
		if fr == nil {
			t.Fatal("expected function response")
		}
		errStr, ok := fr.Response["error"].(string)
		if !ok || !strings.Contains(errStr, "tool execution timed out") {
			t.Errorf("expected timeout error in response, got: %v", fr.Response)
		}
	})

	t.Run("Fallback to Parent Context Deadline when ToolTimeout is zero", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
		defer cancel()

		invCtx := icontext.NewInvocationContext(timeoutCtx, icontext.InvocationContextParams{
			InvocationID: "inv_timeout_3",
			Agent:        &mockAgent{name: "agent_3"},
		})

		flow := &Flow{Tools: []tool.Tool{slowSyncTool}}
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{FunctionCall: &genai.FunctionCall{ID: "c3", Name: "slow_sync", Args: map[string]any{}}},
				},
			},
		}

		mergedEvent, err := flow.handleFunctionCalls(invCtx, toolsDict, resp, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mergedEvent == nil || len(mergedEvent.LLMResponse.Content.Parts) != 1 {
			t.Fatalf("invalid merged event: %#v", mergedEvent)
		}

		fr := mergedEvent.LLMResponse.Content.Parts[0].FunctionResponse
		if fr == nil {
			t.Fatal("expected function response")
		}
		errStr, ok := fr.Response["error"].(string)
		if !ok || !strings.Contains(errStr, "tool execution timed out") {
			t.Errorf("expected timeout error in response, got: %v", fr.Response)
		}
	})
}

type mockFailingLLM struct {
	model.LLM
}

func (m *mockFailingLLM) Name() string { return "failing-model" }

func (m *mockFailingLLM) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		if !yield(&model.LLMResponse{Content: &genai.Content{Parts: []*genai.Part{{Text: "chunk1"}}}}, nil) {
			return
		}
		yield(nil, errors.New("model stream failure"))
	}
}

func TestCallLLM_StreamAndCallbackErrorTeardown(t *testing.T) {
	t.Run("model stream error halts iteration immediately", func(t *testing.T) {
		f := &Flow{Model: &mockFailingLLM{}}
		invCtx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{
			InvocationID: "inv_err_1",
			Agent:        &mockAgent{name: "agent_1"},
		})
		req := &model.LLMRequest{Model: f.Model.Name()}

		var receivedChunks int
		var gotErr error
		for resp, err := range f.callLLM(invCtx, req, make(map[string]any), make(map[string]int64)) {
			if err != nil {
				gotErr = err
				break
			}
			if resp != nil {
				receivedChunks++
			}
		}

		if receivedChunks != 1 {
			t.Errorf("receivedChunks = %d, want 1", receivedChunks)
		}
		if gotErr == nil || gotErr.Error() != "model stream failure" {
			t.Errorf("expected 'model stream failure', got %v", gotErr)
		}
	})

	t.Run("after model callback error halts stream immediately", func(t *testing.T) {
		f := &Flow{
			Model: &mockLLMWithTools{name: "ok-model"},
			AfterModelCallbacks: []AfterModelCallback{
				func(ctx agent.Context, llmResponse *model.LLMResponse, llmResponseError error) (*model.LLMResponse, error) {
					return nil, errors.New("after model callback error")
				},
			},
		}
		invCtx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{
			InvocationID: "inv_err_2",
			Agent:        &mockAgent{name: "agent_2"},
		})
		req := &model.LLMRequest{Model: f.Model.Name()}

		var gotErr error
		for _, err := range f.callLLM(invCtx, req, make(map[string]any), make(map[string]int64)) {
			if err != nil {
				gotErr = err
				break
			}
		}

		if gotErr == nil || gotErr.Error() != "after model callback error" {
			t.Errorf("expected 'after model callback error', got %v", gotErr)
		}
	})
}
