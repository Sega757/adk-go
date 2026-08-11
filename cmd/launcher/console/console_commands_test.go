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

func TestHandleSlashCommand(t *testing.T) {
	tests := []struct {
		name        string
		userInput   string
		tty         bool
		wantHandled bool
		wantExit    bool
		wantOutput  []string
	}{
		{
			name:        "not a slash command",
			userInput:   "hello",
			tty:         false,
			wantHandled: false,
			wantExit:    false,
		},
		{
			name:        "exit command plain",
			userInput:   "/exit",
			tty:         false,
			wantHandled: true,
			wantExit:    true,
			wantOutput:  []string{"Goodbye!"},
		},
		{
			name:        "exit command TTY styled",
			userInput:   "/exit",
			tty:         true,
			wantHandled: true,
			wantExit:    true,
			wantOutput:  []string{"\033[1;32mGoodbye! 👋\033[0m"},
		},
		{
			name:        "quit command plain case-insensitive",
			userInput:   "/QUIT",
			tty:         false,
			wantHandled: true,
			wantExit:    true,
			wantOutput:  []string{"Goodbye!"},
		},
		{
			name:        "clear command plain",
			userInput:   "/clear",
			tty:         false,
			wantHandled: true,
			wantExit:    false,
			wantOutput:  []string{},
		},
		{
			name:        "clear command TTY",
			userInput:   "/clear",
			tty:         true,
			wantHandled: true,
			wantExit:    false,
			wantOutput:  []string{"\033[H\033[2J"},
		},
		{
			name:        "help command plain",
			userInput:   "/help",
			tty:         false,
			wantHandled: true,
			wantExit:    false,
			wantOutput:  []string{"Available Console Commands:", "/help", "/clear", "/exit", "/quit"},
		},
		{
			name:        "help command TTY",
			userInput:   "/help",
			tty:         true,
			wantHandled: true,
			wantExit:    false,
			wantOutput:  []string{"Available Console Commands:", "\033[1;32m/help\033[0m"},
		},
		{
			name:        "invalid command plain",
			userInput:   "/invalid",
			tty:         false,
			wantHandled: true,
			wantExit:    false,
			wantOutput:  []string{"Unknown command: /invalid", "Type /help to see available commands."},
		},
		{
			name:        "invalid command TTY",
			userInput:   "/invalid",
			tty:         true,
			wantHandled: true,
			wantExit:    false,
			wantOutput:  []string{"\033[1;31mUnknown command: /invalid\033[0m", "Type \033[1;32m/help\033[0m"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				handled, exit := handleSlashCommand(tc.userInput, tc.tty)
				if handled != tc.wantHandled {
					t.Errorf("handleSlashCommand() handled = %v, want %v", handled, tc.wantHandled)
				}
				if exit != tc.wantExit {
					t.Errorf("handleSlashCommand() exit = %v, want %v", exit, tc.wantExit)
				}
			})

			for _, expected := range tc.wantOutput {
				if !strings.Contains(out, expected) {
					t.Errorf("expected output to contain %q, but got %q", expected, out)
				}
			}
		})
	}
}
