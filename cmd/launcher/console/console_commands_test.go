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

package console

import (
	"strings"
	"testing"
)

func TestHandleSlashCommand(t *testing.T) {
	tests := []struct {
		name        string
		userInput   string
		tty         bool
		wantExit    bool
		wantHandled bool
		wantStdout  []string
	}{
		{
			name:        "not a slash command",
			userInput:   "hello world",
			tty:         false,
			wantExit:    false,
			wantHandled: false,
			wantStdout:  nil,
		},
		{
			name:        "exit command non-TTY",
			userInput:   "/exit",
			tty:         false,
			wantExit:    true,
			wantHandled: true,
			wantStdout:  []string{"Goodbye!\n"},
		},
		{
			name:        "exit command TTY",
			userInput:   "/exit",
			tty:         true,
			wantExit:    true,
			wantHandled: true,
			wantStdout:  []string{"\033[1;33mGoodbye! 👋\033[0m\n"},
		},
		{
			name:        "quit command TTY",
			userInput:   "/quit",
			tty:         true,
			wantExit:    true,
			wantHandled: true,
			wantStdout:  []string{"\033[1;33mGoodbye! 👋\033[0m\n"},
		},
		{
			name:        "clear command non-TTY",
			userInput:   "/clear",
			tty:         false,
			wantExit:    false,
			wantHandled: true,
			wantStdout:  []string{"[Screen Cleared]\n"},
		},
		{
			name:        "clear command TTY",
			userInput:   "/clear",
			tty:         true,
			wantExit:    false,
			wantHandled: true,
			wantStdout:  []string{"\033[H\033[2J"},
		},
		{
			name:        "help command non-TTY",
			userInput:   "/help",
			tty:         false,
			wantExit:    false,
			wantHandled: true,
			wantStdout:  []string{"Available commands:", "  /help  - Show this help message", "  /clear - Clear the terminal screen"},
		},
		{
			name:        "help command TTY",
			userInput:   "/help",
			tty:         true,
			wantExit:    false,
			wantHandled: true,
			wantStdout:  []string{"\033[1;36mAvailable commands:\033[0m", "  \033[1;32m/help\033[0m", "  \033[1;32m/clear\033[0m"},
		},
		{
			name:        "unrecognized command is NOT intercepted (passed through)",
			userInput:   "/foo",
			tty:         false,
			wantExit:    false,
			wantHandled: false,
			wantStdout:  nil,
		},
		{
			name:        "absolute path is NOT intercepted (passed through)",
			userInput:   "/usr/bin/go",
			tty:         false,
			wantExit:    false,
			wantHandled: false,
			wantStdout:  nil,
		},
		{
			name:        "help command with invalid options non-TTY",
			userInput:   "/help option1 option2",
			tty:         false,
			wantExit:    false,
			wantHandled: true,
			wantStdout:  []string{"Unknown option: option1 option2. Type /help for available commands.\n"},
		},
		{
			name:        "help command with invalid options TTY",
			userInput:   "/help invalid",
			tty:         true,
			wantExit:    false,
			wantHandled: true,
			wantStdout:  []string{"\033[1;31mUnknown option: invalid. Type /help for available commands.\033[0m\n"},
		},
		{
			name:        "clear command with invalid options",
			userInput:   "/clear extra_arg",
			tty:         false,
			wantExit:    false,
			wantHandled: true,
			wantStdout:  []string{"Unknown option: extra_arg. Type /help for available commands.\n"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotExit, gotHandled bool
			out := captureStdout(t, func() {
				gotExit, gotHandled = handleSlashCommand(tc.userInput, tc.tty)
			})

			if gotExit != tc.wantExit {
				t.Errorf("handleSlashCommand exit = %v, want %v", gotExit, tc.wantExit)
			}
			if gotHandled != tc.wantHandled {
				t.Errorf("handleSlashCommand handled = %v, want %v", gotHandled, tc.wantHandled)
			}

			for _, expectedSub := range tc.wantStdout {
				if !strings.Contains(out, expectedSub) {
					t.Errorf("expected stdout to contain %q, but got %q", expectedSub, out)
				}
			}
		})
	}
}
