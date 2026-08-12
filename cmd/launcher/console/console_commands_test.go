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
	"bytes"
	"strings"
	"testing"
)

func TestHandleSlashCommand(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		tty            bool
		wantHandled    bool
		wantShouldExit bool
		wantContains   []string
	}{
		{
			name:           "not a slash command",
			input:          "hello world",
			tty:            false,
			wantHandled:    false,
			wantShouldExit: false,
			wantContains:   nil,
		},
		{
			name:           "empty slash input",
			input:          "/",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"Unrecognized command: /"},
		},
		{
			name:           "exit non-tty",
			input:          "/exit",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: true,
			wantContains:   []string{"Goodbye!"},
		},
		{
			name:           "exit tty",
			input:          "/exit",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: true,
			wantContains:   []string{"\033[1;33mGoodbye! 👋\033[0m"},
		},
		{
			name:           "quit non-tty",
			input:          "/quit",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: true,
			wantContains:   []string{"Goodbye!"},
		},
		{
			name:           "quit tty",
			input:          "/quit",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: true,
			wantContains:   []string{"\033[1;33mGoodbye! 👋\033[0m"},
		},
		{
			name:           "clear non-tty",
			input:          "/clear",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"--- Screen Cleared ---"},
		},
		{
			name:           "clear tty",
			input:          "/clear",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"\033[H\033[2J"},
		},
		{
			name:           "help non-tty",
			input:          "/help",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"Available Console Commands:", "/help", "/clear", "/exit", "/quit"},
		},
		{
			name:           "help tty",
			input:          "/help",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"\033[1;36mAvailable Console Commands:\033[0m", "\033[1;32m/help\033[0m"},
		},
		{
			name:           "invalid command non-tty",
			input:          "/invalid",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"Error: Unrecognized command: /invalid", "Type /help to see available commands."},
		},
		{
			name:           "invalid command tty",
			input:          "/invalid",
			tty:            true,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"\033[1;31mError ❌ -> Unrecognized command: /invalid\033[0m", "\033[1;32m/help\033[0m"},
		},
		{
			name:           "command with arguments is still processed",
			input:          "/help extra_args",
			tty:            false,
			wantHandled:    true,
			wantShouldExit: false,
			wantContains:   []string{"Available Console Commands:"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			handled, shouldExit := handleSlashCommand(tc.input, tc.tty, &buf)
			if handled != tc.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tc.wantHandled)
			}
			if shouldExit != tc.wantShouldExit {
				t.Errorf("shouldExit = %v, want %v", shouldExit, tc.wantShouldExit)
			}
			out := buf.String()
			for _, s := range tc.wantContains {
				if !strings.Contains(out, s) {
					t.Errorf("expected output to contain %q, but got:\n%s", s, out)
				}
			}
		})
	}
}
