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

package console

import (
	"strings"
	"testing"
)

func TestHandleConsoleCommand(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		tty            bool
		wantHandled    bool
		wantShouldExit bool
		wantSubstrings []string
	}{
		{
			name:           "help command non-tty",
			input:          "/help",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: false,
			wantSubstrings: []string{"Available Console Commands:", "/help", "/clear", "/exit"},
		},
		{
			name:           "help command tty",
			input:          " /help ",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: false,
			wantSubstrings: []string{"Available Console Commands", "/help", "/clear", "/exit"},
		},
		{
			name:           "clear command non-tty",
			input:          "/clear",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: false,
			wantSubstrings: []string{"[Screen Cleared]"},
		},
		{
			name:           "clear command tty",
			input:          "/clear",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: false,
			wantSubstrings: []string{"\033[H\033[2J"},
		},
		{
			name:           "exit command non-tty",
			input:          "/exit",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: true,
			wantSubstrings: []string{"Goodbye!"},
		},
		{
			name:           "quit command tty",
			input:          "/quit",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: true,
			wantSubstrings: []string{"Goodbye! 👋"},
		},
		{
			name:           "non-command slash path passthrough",
			input:          "/tmp/file.txt",
			tty:            true,
			wantHandled:    false,
			wantShouldExit: false,
			wantSubstrings: []string{},
		},
		{
			name:           "regular message passthrough",
			input:          "Hello Agent",
			tty:            true,
			wantHandled:    false,
			wantShouldExit: false,
			wantSubstrings: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handled, shouldExit bool
			out := captureStdout(t, func() {
				handled, shouldExit = handleConsoleCommand(tt.input, tt.tty)
			})

			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
			if shouldExit != tt.wantShouldExit {
				t.Errorf("shouldExit = %v, want %v", shouldExit, tt.wantShouldExit)
			}
			for _, sub := range tt.wantSubstrings {
				if !strings.Contains(out, sub) {
					t.Errorf("output %q does not contain expected substring %q", out, sub)
				}
			}
		})
	}
}
