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

func TestConsoleCommands_HelpMessage(t *testing.T) {
	t.Run("Non-TTY mode", func(t *testing.T) {
		out := captureStdout(t, func() {
			printHelpMessage(false)
		})
		if !strings.Contains(out, "Console Commands:") {
			t.Errorf("expected plain help title, got: %q", out)
		}
		if !strings.Contains(out, "/help") || !strings.Contains(out, "/clear") || !strings.Contains(out, "/exit") {
			t.Errorf("expected help commands list, got: %q", out)
		}
		if strings.Contains(out, "\033") {
			t.Errorf("expected non-TTY help output to not contain ANSI escape codes, got: %q", out)
		}
	})

	t.Run("TTY mode", func(t *testing.T) {
		out := captureStdout(t, func() {
			printHelpMessage(true)
		})
		if !strings.Contains(out, "💡 Console Commands:") {
			t.Errorf("expected TTY help title with emoji, got: %q", out)
		}
		if !strings.Contains(out, "\033[1;35m") || !strings.Contains(out, "\033[1;36m") {
			t.Errorf("expected ANSI color codes in TTY mode, got: %q", out)
		}
	})
}

func TestConsoleCommands_ClearScreen(t *testing.T) {
	t.Run("Non-TTY mode", func(t *testing.T) {
		out := captureStdout(t, func() {
			printClearScreen(false)
		})
		if !strings.Contains(out, "Screen Cleared") {
			t.Errorf("expected non-TTY clear screen notice, got: %q", out)
		}
	})

	t.Run("TTY mode", func(t *testing.T) {
		out := captureStdout(t, func() {
			printClearScreen(true)
		})
		if !strings.Contains(out, "\033[H\033[2J") {
			t.Errorf("expected TTY ANSI clear screen escape sequence, got: %q", out)
		}
	})
}
