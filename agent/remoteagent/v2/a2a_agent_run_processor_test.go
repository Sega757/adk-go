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

package remoteagent

import (
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	icontext "google.golang.org/adk/v2/internal/context"
	"google.golang.org/adk/v2/internal/utils"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/server/adka2a/v2"
	"google.golang.org/adk/v2/session"
)

func TestA2AAgentRunProcessor_aggregatePartial(t *testing.T) {
	type updateFlags struct {
		append    bool
		lastChunk bool
	}
	task := &a2a.Task{ID: a2a.NewTaskID(), ContextID: a2a.NewContextID()}
	newArtifactUpdate := func(aid a2a.ArtifactID, flags updateFlags, text string) *a2a.TaskArtifactUpdateEvent {
		return &a2a.TaskArtifactUpdateEvent{
			TaskID:    task.ID,
			ContextID: task.ContextID,
			Artifact:  &a2a.Artifact{ID: aid, Parts: a2a.ContentParts{a2a.NewTextPart(text)}},
			LastChunk: flags.lastChunk,
			Append:    flags.append,
		}
	}

	newPartialEvent := func(text string) *session.Event {
		return &session.Event{LLMResponse: model.LLMResponse{
			Partial: true,
			Content: genai.NewContentFromText(text, genai.RoleModel),
		}}
	}
	newCompletedEvent := func(parts ...*genai.Part) *session.Event {
		e := &session.Event{LLMResponse: model.LLMResponse{TurnComplete: true}}
		if len(parts) > 0 {
			e.Content = genai.NewContentFromParts(parts, genai.RoleModel)
		}
		return e
	}
	newEvent := func(parts ...*genai.Part) *session.Event {
		e := &session.Event{LLMResponse: model.LLMResponse{Partial: false}}
		if len(parts) > 0 {
			e.Content = genai.NewContentFromParts(parts, genai.RoleModel)
		}
		return e
	}
	withADKPartial := func(event *a2a.TaskArtifactUpdateEvent, partial bool) *a2a.TaskArtifactUpdateEvent {
		event.Metadata = map[string]any{adka2a.ToA2AMetaKey("partial"): partial}
		return event
	}

	aid1, aid2 := a2a.NewArtifactID(), a2a.NewArtifactID()
	tests := []struct {
		name       string
		events     []a2a.Event
		wantEvents []*session.Event
	}{
		{
			name: "do not aggregate when ADK events",
			events: []a2a.Event{
				withADKPartial(a2a.NewArtifactUpdateEvent(task, aid1, a2a.NewTextPart("Hel")), true),
				withADKPartial(a2a.NewArtifactUpdateEvent(task, aid1, a2a.NewTextPart("lo")), true),
				withADKPartial(a2a.NewArtifactUpdateEvent(task, aid1, a2a.NewTextPart("Hello")), false),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("Hel"),
				newPartialEvent("lo"),
				newEvent(genai.NewPartFromText("Hello")),
				newCompletedEvent(),
			},
		},
		{
			name: "aggregation reset by final snapshot",
			events: []a2a.Event{
				a2a.NewArtifactUpdateEvent(task, "a1", a2a.NewTextPart("ignore me")),
				&a2a.Task{
					ID:        task.ID,
					Artifacts: []*a2a.Artifact{{Parts: a2a.ContentParts{a2a.NewTextPart("done")}}},
					Status:    a2a.TaskStatus{State: a2a.TaskStateCompleted},
				},
			},
			wantEvents: []*session.Event{
				newPartialEvent("ignore me"),
				newCompletedEvent(genai.NewPartFromText("done")),
			},
		},
		{
			name: "aggregation reset by non-final snapshot",
			events: []a2a.Event{
				a2a.NewArtifactUpdateEvent(task, "a1", a2a.NewTextPart("foo")),
				&a2a.Task{ID: task.ID},
				a2a.NewArtifactUpdateEvent(task, "a1", a2a.NewTextPart("bar")),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("foo"),
				newPartialEvent("bar"),
				newEvent(genai.NewPartFromText("bar")),
				newCompletedEvent(),
			},
		},
		{
			name: "[append=true, lastChunk=false] emit aggregated on final status",
			events: []a2a.Event{
				newArtifactUpdate(aid1, updateFlags{append: true}, "Hel"),
				newArtifactUpdate(aid1, updateFlags{append: true}, "lo"),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("Hel"),
				newPartialEvent("lo"),
				newEvent(genai.NewPartFromText("Hello")),
				newCompletedEvent(),
			},
		},
		{
			name: "[append=true, lastChunk=false] emit multiple aggregated on final status",
			events: []a2a.Event{
				newArtifactUpdate(aid1, updateFlags{append: true}, "Foo"),
				newArtifactUpdate(aid2, updateFlags{append: true}, "Bar"),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("Foo"),
				newPartialEvent("Bar"),
				newEvent(genai.NewPartFromText("Foo")),
				newEvent(genai.NewPartFromText("Bar")),
				newCompletedEvent(),
			},
		},
		{
			name: "last updated aggregation is emitted last",
			events: []a2a.Event{
				newArtifactUpdate(aid1, updateFlags{append: true}, "Foo"),
				newArtifactUpdate(aid2, updateFlags{append: true}, "Bar"),
				newArtifactUpdate(aid1, updateFlags{append: true}, "Baz"),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("Foo"),
				newPartialEvent("Bar"),
				newPartialEvent("Baz"),
				newEvent(genai.NewPartFromText("Bar")),
				newEvent(genai.NewPartFromText("FooBaz")),
				newCompletedEvent(),
			},
		},
		{
			name: "[append=true, lastChunk=true] results in partial followed by non-partial",
			events: []a2a.Event{
				newArtifactUpdate(aid1, updateFlags{append: true}, "Hel"),
				newArtifactUpdate(aid1, updateFlags{append: true, lastChunk: true}, "lo"),
				newArtifactUpdate(aid2, updateFlags{append: true}, "bar"),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("Hel"),
				newPartialEvent("lo"),
				newEvent(genai.NewPartFromText("Hello")),
				newPartialEvent("bar"),
				newEvent(genai.NewPartFromText("bar")),
				newCompletedEvent(),
			},
		},
		{
			name: "[append=false, lastChunk=true] results in non-partial",
			events: []a2a.Event{
				newArtifactUpdate(aid1, updateFlags{append: true}, "Hel"),
				newArtifactUpdate(aid1, updateFlags{append: false, lastChunk: true}, "Hello"),
				newArtifactUpdate(aid2, updateFlags{append: true}, "bar"),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("Hel"),
				newEvent(genai.NewPartFromText("Hello")),
				newPartialEvent("bar"),
				newEvent(genai.NewPartFromText("bar")),
				newCompletedEvent(),
			},
		},
		{
			name: "[append=false, lastChunk=true] as first event non-partial",
			events: []a2a.Event{
				newArtifactUpdate(aid1, updateFlags{append: false, lastChunk: true}, "Hello"),
				newArtifactUpdate(aid2, updateFlags{append: true}, "bar"),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newEvent(genai.NewPartFromText("Hello")),
				newPartialEvent("bar"),
				newEvent(genai.NewPartFromText("bar")),
				newCompletedEvent(),
			},
		},
		{
			name: "[append=false, lastChunk=false] resets aggregation",
			events: []a2a.Event{
				newArtifactUpdate(aid1, updateFlags{append: true}, "Foo"),
				newArtifactUpdate(aid1, updateFlags{append: false}, "Bar"),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				newPartialEvent("Foo"),
				newPartialEvent("Bar"),
				newEvent(genai.NewPartFromText("Bar")),
				newCompletedEvent(),
			},
		},
		{
			name: "thoughts aggregation",
			events: []a2a.Event{
				func() *a2a.TaskArtifactUpdateEvent {
					p := a2a.NewTextPart("thinking...")
					p.SetMeta(adka2a.ToA2AMetaKey("thought"), true)
					return a2a.NewArtifactUpdateEvent(task, "a1", p)
				}(),
				a2a.NewArtifactUpdateEvent(task, "a1", a2a.NewTextPart("done")),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				{LLMResponse: model.LLMResponse{
					Partial: true,
					Content: &genai.Content{Parts: []*genai.Part{{Thought: true, Text: "thinking..."}}, Role: genai.RoleModel},
				}},
				newPartialEvent("done"),
				newEvent(
					&genai.Part{Thought: true, Text: "thinking..."},
					&genai.Part{Text: "done"},
				),
				newCompletedEvent(),
			},
		},
		{
			name: "interleaved thought and text",
			events: []a2a.Event{
				a2a.NewArtifactUpdateEvent(task, "a1", &a2a.Part{
					Content:  a2a.Text("thinking1"),
					Metadata: map[string]any{adka2a.ToA2AMetaKey("thought"): true},
				}),
				a2a.NewArtifactUpdateEvent(task, "a1", a2a.NewTextPart("text1")),
				a2a.NewArtifactUpdateEvent(task, "a1", &a2a.Part{
					Content:  a2a.Text("thinking2"),
					Metadata: map[string]any{adka2a.ToA2AMetaKey("thought"): true},
				}),
				a2a.NewArtifactUpdateEvent(task, "a1", a2a.NewTextPart("text2")),
				newFinalStatusUpdate(task, a2a.TaskStateCompleted),
			},
			wantEvents: []*session.Event{
				{LLMResponse: model.LLMResponse{
					Partial: true,
					Content: &genai.Content{Parts: []*genai.Part{{Thought: true, Text: "thinking1"}}, Role: genai.RoleModel},
				}},
				newPartialEvent("text1"),
				{LLMResponse: model.LLMResponse{
					Partial: true,
					Content: &genai.Content{Parts: []*genai.Part{{Thought: true, Text: "thinking2"}}, Role: genai.RoleModel},
				}},
				newPartialEvent("text2"),
				newEvent(
					&genai.Part{Thought: true, Text: "thinking1"},
					&genai.Part{Text: "text1"},
					&genai.Part{Thought: true, Text: "thinking2"},
					&genai.Part{Text: "text2"},
				),
				newCompletedEvent(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agnt := utils.Must(agent.New(agent.Config{}))
			ctx := icontext.NewInvocationContext(t.Context(), icontext.InvocationContextParams{
				Agent: agnt,
			})

			p := newRunProcessor(A2AConfig{}, nil)
			var gotEvents []*session.Event

			for _, event := range tc.events {

				adkEvent, err := adka2a.ToSessionEvent(ctx, event)
				if err != nil {
					t.Fatalf("ToSessionEvent failed: %v", err)
				}

				if adkEvent == nil {
					// Handle Task snapshot resetting aggregation even if it doesn't produce an event
					if _, ok := event.(*a2a.Task); ok {
						p.aggregatePartial(ctx, event, nil)
					}
					continue
				}

				gotEvents = append(gotEvents, p.aggregatePartial(ctx, event, adkEvent)...)
			}
			opts := []cmp.Option{
				cmpopts.IgnoreFields(session.Event{}, "ID", "Timestamp", "InvocationID", "Author", "Branch", "CustomMetadata", "Actions"),
			}
			if diff := cmp.Diff(tc.wantEvents, gotEvents, opts...); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkArtifactAggregation(b *testing.B) {
	makeEvents := func(n int, alternateThought bool) []*session.Event {
		events := make([]*session.Event, n)
		for i := 0; i < n; i++ {
			thought := false
			if alternateThought {
				thought = (i%2 == 1)
			}
			events[i] = &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{Text: "some chunk of text for aggregation ", Thought: thought},
						},
					},
				},
			}
		}
		return events
	}

	scenarios := []struct {
		name             string
		n                int
		alternateThought bool
	}{
		{name: "Contiguous_1000", n: 1000, alternateThought: false},
		{name: "Contiguous_10000", n: 10000, alternateThought: false},
		{name: "Alternating_1000", n: 1000, alternateThought: true},
		{name: "Alternating_10000", n: 10000, alternateThought: true},
	}

	for _, sc := range scenarios {
		events := makeEvents(sc.n, sc.alternateThought)
		b.Run(sc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p := newRunProcessor(A2AConfig{}, nil)
				agg := &artifactAggregation{}
				aid := a2a.ArtifactID("art-1")
				p.aggregations[aid] = agg
				for _, ev := range events {
					p.updateAggregation(aid, agg, ev)
				}
			}
		})
	}
}

func TestUpdateAggregation_EdgeCases(t *testing.T) {
	t.Run("empty parts and empty text parts", func(t *testing.T) {
		p := newRunProcessor(A2AConfig{}, nil)
		agg := &artifactAggregation{}
		aid := a2a.ArtifactID("art-edge-1")
		p.aggregations[aid] = agg

		emptyTextPart := &genai.Part{Text: ""}
		ev := &session.Event{
			LLMResponse: model.LLMResponse{
				Content: &genai.Content{
					Parts: []*genai.Part{
						emptyTextPart,
						{Text: "chunk 1", Thought: false},
						emptyTextPart,
						{Text: "chunk 2", Thought: false},
					},
				},
			},
		}

		p.updateAggregation(aid, agg, ev)

		if len(agg.parts) != 4 {
			t.Fatalf("expected 4 aggregated parts, got %d", len(agg.parts))
		}
		if agg.parts[0] != emptyTextPart {
			t.Errorf("expected emptyTextPart at index 0, got %v", agg.parts[0])
		}
		if agg.parts[1].Text != "chunk 1" {
			t.Errorf("expected 'chunk 1', got %q", agg.parts[1].Text)
		}
		if agg.parts[2] != emptyTextPart {
			t.Errorf("expected emptyTextPart at index 2, got %v", agg.parts[2])
		}
		if agg.parts[3].Text != "chunk 2" {
			t.Errorf("expected 'chunk 2', got %q", agg.parts[3].Text)
		}
	})

	t.Run("mixed content types and single non-merging parts", func(t *testing.T) {
		p := newRunProcessor(A2AConfig{}, nil)
		agg := &artifactAggregation{}
		aid := a2a.ArtifactID("art-edge-2")
		p.aggregations[aid] = agg

		mediaPart := &genai.Part{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte("img")}}
		ev := &session.Event{
			LLMResponse: model.LLMResponse{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "text A", Thought: false},
						mediaPart,
						{Text: "text B", Thought: false},
						{Text: "text C", Thought: false},
					},
				},
			},
		}

		p.updateAggregation(aid, agg, ev)

		if len(agg.parts) != 3 {
			t.Fatalf("expected 3 aggregated parts, got %d", len(agg.parts))
		}
		if agg.parts[0].Text != "text A" {
			t.Errorf("expected 'text A', got %q", agg.parts[0].Text)
		}
		if agg.parts[1] != mediaPart {
			t.Errorf("expected mediaPart at index 1, got %v", agg.parts[1])
		}
		if agg.parts[2].Text != "text Btext C" {
			t.Errorf("expected 'text Btext C', got %q", agg.parts[2].Text)
		}
	})

	t.Run("nil and empty content events", func(t *testing.T) {
		p := newRunProcessor(A2AConfig{}, nil)
		agg := &artifactAggregation{}
		aid := a2a.ArtifactID("art-edge-3")
		p.aggregations[aid] = agg

		evNil := &session.Event{}
		evEmptyContent := &session.Event{LLMResponse: model.LLMResponse{Content: &genai.Content{}}}

		p.updateAggregation(aid, agg, evNil)
		p.updateAggregation(aid, agg, evEmptyContent)

		if len(agg.parts) != 0 {
			t.Fatalf("expected 0 aggregated parts, got %d", len(agg.parts))
		}
	})
}
