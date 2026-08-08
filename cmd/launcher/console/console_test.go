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
	"iter"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/session"
)

func runConsoleTest(t *testing.T, input string) (string, error) {
	t.Helper()
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r

	if _, err := w.Write([]byte(input)); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}
	w.Close()

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = outW

	outputChan := make(chan string)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, outR)
		outputChan <- buf.String()
	}()

	agnt, err := agent.New(agent.Config{
		Name: "dummy",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {}
		},
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader:    agent.NewSingleLoader(agnt),
		SessionService: session.InMemoryService(),
	}

	cl := NewLauncher().(*consoleLauncher)
	cl.config.shutdownTimeout = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	runErr := cl.Run(ctx, config)

	os.Stdout = oldStdout
	os.Stdin = oldStdin
	outW.Close()

	outStr := <-outputChan
	return outStr, runErr
}

func TestConsole_SlashExit(t *testing.T) {
	out, err := runConsoleTest(t, "/exit\n")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "Goodbye! 👋") {
		t.Errorf("expected output to contain Goodbye message, got %q", out)
	}
}

func TestConsole_SlashQuit(t *testing.T) {
	out, err := runConsoleTest(t, "/quit\n")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "Goodbye! 👋") {
		t.Errorf("expected output to contain Goodbye message, got %q", out)
	}
}

func TestConsole_SlashHelp(t *testing.T) {
	out, err := runConsoleTest(t, "/help\n")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "Console Commands:") {
		t.Errorf("expected output to contain command list, got %q", out)
	}
	if !strings.Contains(out, "/help") || !strings.Contains(out, "/clear") || !strings.Contains(out, "/exit") {
		t.Errorf("expected commands in list, got %q", out)
	}
}

func TestConsole_SlashClear(t *testing.T) {
	out, err := runConsoleTest(t, "/clear\n")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "--- Screen Cleared ---") {
		t.Errorf("expected output to clear screen, got %q", out)
	}
}
