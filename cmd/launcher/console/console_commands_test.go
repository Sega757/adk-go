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

func TestPrintHelpPrompt(t *testing.T) {
	t.Run("Non-TTY mode", func(t *testing.T) {
		out := captureStdout(t, func() {
			printHelpPrompt(false)
		})
		if !strings.Contains(out, "--- ADK Console Help ---") {
			t.Errorf("expected header in plain help output, got %q", out)
		}
		if !strings.Contains(out, "/help") || !strings.Contains(out, "/clear") || !strings.Contains(out, "/exit") {
			t.Errorf("expected slash commands listed in plain help output, got %q", out)
		}
		if strings.Contains(out, "\033") {
			t.Errorf("plain help output should not contain ANSI escape codes, got %q", out)
		}
	})

	t.Run("TTY mode", func(t *testing.T) {
		out := captureStdout(t, func() {
			printHelpPrompt(true)
		})
		if !strings.Contains(out, "\033[1;36m--- ADK Console Help ---\033[0m") {
			t.Errorf("expected ANSI styled header in TTY help output, got %q", out)
		}
		if !strings.Contains(out, "/help") || !strings.Contains(out, "/clear") || !strings.Contains(out, "/exit") {
			t.Errorf("expected slash commands listed in TTY help output, got %q", out)
		}
	})
}
