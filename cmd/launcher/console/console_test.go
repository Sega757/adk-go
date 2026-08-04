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
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/session"
)

func TestConsoleLauncher_Commands(t *testing.T) {
	tests := []struct {
		name         string
		inputs       []string
		wantStdout   []string
		wantNoStdout []string
	}{
		{
			name:   "exit command cleanly exits",
			inputs: []string{"/exit\n"},
			wantStdout: []string{
				"Goodbye! 👋",
			},
		},
		{
			name:   "quit command cleanly exits",
			inputs: []string{"/quit\n"},
			wantStdout: []string{
				"Goodbye! 👋",
			},
		},
		{
			name:   "help command prints help instructions",
			inputs: []string{"/help\n", "/exit\n"},
			wantStdout: []string{
				"Console Commands",
				"/help",
				"/clear",
				"/exit",
			},
		},
		{
			name:   "clear command clears terminal and re-prints welcome on TTY",
			inputs: []string{"/clear\n", "/exit\n"},
			wantStdout: []string{
				"Welcome to ADK Console",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			// Mock Stdin using os.Pipe
			inR, inW, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			origStdin := os.Stdin
			os.Stdin = inR
			defer func() { os.Stdin = origStdin }()

			// Write input in a separate goroutine so it doesn't block
			go func() {
				for _, input := range tc.inputs {
					_, _ = inW.Write([]byte(input))
				}
				_ = inW.Close()
			}()

			// Redirect Stdout to capture output
			origStdout := os.Stdout
			outR, outW, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stdout = outW

			// Run launcher
			l := NewLauncher().(*consoleLauncher)
			// Configure to avoid telemetry network calls or delays
			l.config.shutdownTimeout = 10 * time.Millisecond

			cfg := &launcher.Config{
				SessionService: session.InMemoryService(),
				AgentLoader:    agent.NewSingleLoader(newConsoleHITLAgent(t)),
			}

			err = l.Run(ctx, cfg)
			_ = outW.Close()
			os.Stdout = origStdout

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			stdoutBytes, _ := io.ReadAll(outR)
			stdoutStr := string(stdoutBytes)

			for _, want := range tc.wantStdout {
				if !strings.Contains(stdoutStr, want) {
					t.Errorf("expected stdout to contain %q, but got:\n%s", want, stdoutStr)
				}
			}
			for _, wantNo := range tc.wantNoStdout {
				if strings.Contains(stdoutStr, wantNo) {
					t.Errorf("expected stdout NOT to contain %q, but got:\n%s", wantNo, stdoutStr)
				}
			}
		})
	}
}
