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

package replayplugin

import (
	"path/filepath"
	"sync"
	"testing"

	"google.golang.org/adk/v2/internal/configurable/conformance/replayplugin/recording"
)

func TestNewInvocationReplayStateAndGetters(t *testing.T) {
	dummyRecordings := &recording.Recordings{}

	tests := []struct {
		name                 string
		testCasePath         string
		userMessageIndex     int
		recs                 *recording.Recordings
		agentToQuery         string
		initialAgentIndices  map[string]int
		wantAgentIndex       int
		wantAgentFound       bool
	}{
		{
			name:             "Standard initialization with valid path and non-nil recordings",
			testCasePath:     filepath.Join("testdata", "cases", "basic.yaml"),
			userMessageIndex: 2,
			recs:             dummyRecordings,
			agentToQuery:     "agent1",
			wantAgentIndex:   0,
			wantAgentFound:   false,
		},
		{
			name:             "Empty test case path and zero index",
			testCasePath:     "",
			userMessageIndex: 0,
			recs:             nil,
			agentToQuery:     "nonexistent",
			wantAgentIndex:   0,
			wantAgentFound:   false,
		},
		{
			name:             "Path with spaces, dots, and relative separators",
			testCasePath:     "./folder with spaces/../test_case.yaml",
			userMessageIndex: -1,
			recs:             dummyRecordings,
			agentToQuery:     "agent_a",
			initialAgentIndices: map[string]int{
				"agent_a": 5,
			},
			wantAgentIndex: 5,
			wantAgentFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newInvocationReplayState(tt.testCasePath, tt.userMessageIndex, tt.recs)

			if state == nil {
				t.Fatalf("newInvocationReplayState() returned nil state")
			}

			// Verify GetTestCasePath
			if got := state.GetTestCasePath(); got != tt.testCasePath {
				t.Errorf("GetTestCasePath() = %q; want %q", got, tt.testCasePath)
			}

			// Verify GetUserMessageIndex
			if got := state.GetUserMessageIndex(); got != tt.userMessageIndex {
				t.Errorf("GetUserMessageIndex() = %d; want %d", got, tt.userMessageIndex)
			}

			// Verify GetRecordings
			if got := state.GetRecordings(); got != tt.recs {
				t.Errorf("GetRecordings() = %v; want %v", got, tt.recs)
			}

			// Verify map and cond initialization
			if state.agentReplayIndices == nil {
				t.Errorf("agentReplayIndices map was not initialized")
			}
			if state.consumedRecordings == nil {
				t.Errorf("consumedRecordings map was not initialized")
			}
			if state.cond == nil {
				t.Errorf("cond condition variable was not initialized")
			}

			// Pre-populate map if test scenario specifies initial entries
			if tt.initialAgentIndices != nil {
				state.mu.Lock()
				for k, v := range tt.initialAgentIndices {
					state.agentReplayIndices[k] = v
				}
				state.mu.Unlock()
			}

			// Verify GetAgentReplayIndex
			gotIdx, gotOk := state.GetAgentReplayIndex(tt.agentToQuery)
			if gotIdx != tt.wantAgentIndex || gotOk != tt.wantAgentFound {
				t.Errorf("GetAgentReplayIndex(%q) = (%d, %t); want (%d, %t)",
					tt.agentToQuery, gotIdx, gotOk, tt.wantAgentIndex, tt.wantAgentFound)
			}
		})
	}
}

func TestZeroValueInvocationReplayState(t *testing.T) {
	var zeroState invocationReplayState

	if got := zeroState.GetTestCasePath(); got != "" {
		t.Errorf("Zero value GetTestCasePath() = %q; want %q", got, "")
	}

	if got := zeroState.GetUserMessageIndex(); got != 0 {
		t.Errorf("Zero value GetUserMessageIndex() = %d; want 0", got)
	}

	if got := zeroState.GetRecordings(); got != nil {
		t.Errorf("Zero value GetRecordings() = %v; want nil", got)
	}

	// GetAgentReplayIndex on uninitialized map should safely return (0, false)
	idx, ok := zeroState.GetAgentReplayIndex("test_agent")
	if idx != 0 || ok != false {
		t.Errorf("GetAgentReplayIndex on zero value = (%d, %t); want (0, false)", idx, ok)
	}
}

func TestConcurrentGetAgentReplayIndex(t *testing.T) {
	state := newInvocationReplayState("concurrent_test.yaml", 1, nil)

	const numGoroutines = 20
	var wg sync.WaitGroup

	// Concurrently read and write agent replay indices
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			agentName := "agent"
			if id%2 == 0 {
				state.mu.Lock()
				state.agentReplayIndices[agentName] = id
				state.mu.Unlock()
			} else {
				state.GetAgentReplayIndex(agentName)
			}
		}(i)
	}

	wg.Wait()
}
