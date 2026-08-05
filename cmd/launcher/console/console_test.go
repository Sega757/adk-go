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
	"context"
	"iter"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/session"
)

func TestConsoleLauncher_SlashCommands(t *testing.T) {
	tests := []struct {
		name        string
		inputs      []string
		wantOutputs []string
		wantNoExit  bool
	}{
		{
			name:   "help command",
			inputs: []string{"/help\n"},
			wantOutputs: []string{
				"Console Commands:",
				"/help",
				"/clear",
				"/exit",
			},
			wantNoExit: true,
		},
		{
			name:   "clear command",
			inputs: []string{"/clear\n"},
			wantOutputs: []string{
				"Welcome to ADK Console!",
			},
			wantNoExit: true,
		},
		{
			name:   "exit command",
			inputs: []string{"/exit\n"},
			wantOutputs: []string{
				"Goodbye! 👋",
			},
			wantNoExit: false,
		},
		{
			name:   "quit command",
			inputs: []string{"/quit\n"},
			wantOutputs: []string{
				"Goodbye! 👋",
			},
			wantNoExit: false,
		},
		{
			name:   "unknown command fallback",
			inputs: []string{"/invalidcmd\n"},
			wantOutputs: []string{
				"Unknown command: /invalidcmd",
				"Type /help to see available commands",
			},
			wantNoExit: true,
		},
	}

	mockAg, err := agent.New(agent.Config{
		Name: "mock",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {}
		},
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Mock stdin
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			defer r.Close()

			origStdin := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = origStdin }()

			// Write mocked inputs and close writer so launcher loop terminates/receives EOF
			go func() {
				for _, in := range tc.inputs {
					_, _ = w.Write([]byte(in))
				}
				if tc.wantNoExit {
					// send EOF to exit loop if we didn't exit via /exit
					_ = w.Close()
				}
			}()

			l := NewLauncher().(*consoleLauncher)
			// Configure rapid timeout for testing
			l.config.shutdownTimeout = 10 * time.Millisecond

			// Intercept stdout to verify printed output
			out := captureStdout(t, func() {
				config := &launcher.Config{
					SessionService: session.InMemoryService(),
					AgentLoader:    agent.NewSingleLoader(mockAg),
				}
				_ = l.Run(context.Background(), config)
			})

			for _, expected := range tc.wantOutputs {
				if !strings.Contains(out, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, out)
				}
			}
		})
	}
}
