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

func TestHandleSlashCommand_ExitAndQuit(t *testing.T) {
	tests := []struct {
		name    string
		command string
		tty     bool
	}{
		{"exit with tty", "/exit", true},
		{"exit without tty", "/exit", false},
		{"quit with tty", "/quit", true},
		{"quit without tty", "/quit", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			exit, handled := handleSlashCommand(&buf, tc.command, tc.tty)
			if !exit {
				t.Errorf("handleSlashCommand(%q) exit = false, want true", tc.command)
			}
			if !handled {
				t.Errorf("handleSlashCommand(%q) handled = false, want true", tc.command)
			}

			output := buf.String()
			if tc.tty {
				if !strings.Contains(output, "\033[1;33m") {
					t.Errorf("expected TTY-styled output, got %q", output)
				}
			} else {
				if strings.Contains(output, "\033") {
					t.Errorf("expected plain text, got ANSI escape sequences: %q", output)
				}
			}
			if !strings.Contains(output, "Exiting ADK Console...") {
				t.Errorf("output %q does not contain exit message", output)
			}
		})
	}
}

func TestHandleSlashCommand_Clear(t *testing.T) {
	tests := []struct {
		name string
		tty  bool
	}{
		{"clear with tty", true},
		{"clear without tty", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			exit, handled := handleSlashCommand(&buf, "/clear", tc.tty)
			if exit {
				t.Errorf("handleSlashCommand(/clear) exit = true, want false")
			}
			if !handled {
				t.Errorf("handleSlashCommand(/clear) handled = false, want true")
			}

			output := buf.String()
			if !strings.HasPrefix(output, "\033[H\033[2J") {
				t.Errorf("expected clear screen ANSI prefix, got %q", output)
			}

			if tc.tty {
				if !strings.Contains(output, "User 👤 ->") {
					t.Errorf("expected TTY user prompt, got %q", output)
				}
			} else {
				if !strings.Contains(output, "User ->") {
					t.Errorf("expected plain user prompt, got %q", output)
				}
			}
		})
	}
}

func TestHandleSlashCommand_Help(t *testing.T) {
	tests := []struct {
		name string
		tty  bool
	}{
		{"help with tty", true},
		{"help without tty", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			exit, handled := handleSlashCommand(&buf, "/help", tc.tty)
			if exit {
				t.Errorf("handleSlashCommand(/help) exit = true, want false")
			}
			if !handled {
				t.Errorf("handleSlashCommand(/help) handled = false, want true")
			}

			output := buf.String()
			if !strings.Contains(output, "Available commands:") {
				t.Errorf("output %q does not contain 'Available commands:' header", output)
			}
			if !strings.Contains(output, "/help") || !strings.Contains(output, "/clear") || !strings.Contains(output, "/exit") || !strings.Contains(output, "/quit") {
				t.Errorf("output %q is missing some command entries", output)
			}

			if tc.tty {
				if !strings.Contains(output, "\033[1;36m") {
					t.Errorf("expected TTY-styled output, got %q", output)
				}
				if !strings.Contains(output, "User 👤 ->") {
					t.Errorf("expected user prompt with emoji/color, got %q", output)
				}
			} else {
				if strings.Contains(output, "\033[1;36m") {
					t.Errorf("expected plain text, got TTY ANSI sequences: %q", output)
				}
				if !strings.Contains(output, "User ->") {
					t.Errorf("expected plain user prompt, got %q", output)
				}
			}
		})
	}
}

func TestHandleSlashCommand_Unknown(t *testing.T) {
	tests := []struct {
		name    string
		command string
		tty     bool
	}{
		{"unknown command with tty", "/unknown_cmd", true},
		{"unknown command without tty", "/unknown_cmd", false},
		{"slash prefix only with tty", "/", true},
		{"slash prefix only without tty", "/", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			exit, handled := handleSlashCommand(&buf, tc.command, tc.tty)
			if exit {
				t.Errorf("handleSlashCommand(%q) exit = true, want false", tc.command)
			}
			if !handled {
				t.Errorf("handleSlashCommand(%q) handled = false, want true", tc.command)
			}

			output := buf.String()
			if !strings.Contains(output, "Unknown command:") {
				t.Errorf("output %q does not contain unknown command message", output)
			}
			if !strings.Contains(output, tc.command) {
				t.Errorf("output %q does not mention unknown command %q", output, tc.command)
			}

			if tc.tty {
				if !strings.Contains(output, "\033[1;31m") {
					t.Errorf("expected TTY-styled output, got %q", output)
				}
			} else {
				if strings.Contains(output, "\033") {
					t.Errorf("expected plain text, got ANSI escape sequences: %q", output)
				}
			}
		})
	}
}
