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

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

func runConsoleWithInput(t *testing.T, input string) string {
	t.Helper()

	// Redirect standard input
	inReader, inWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create input pipe: %v", err)
	}
	origStdin := os.Stdin
	os.Stdin = inReader
	defer func() { os.Stdin = origStdin }()

	// Redirect standard output
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		inWriter.Close()
		t.Fatalf("failed to create output pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = outWriter
	defer func() { os.Stdout = origStdout }()

	// Write the simulated user inputs
	go func() {
		defer inWriter.Close()
		_, _ = inWriter.Write([]byte(input))
	}()

	// Capture the printed outputs
	outChan := make(chan string, 1)
	go func() {
		defer outReader.Close()
		b, _ := io.ReadAll(outReader)
		outChan <- string(b)
	}()

	// Create a simple test agent
	testAgent, err := agent.New(agent.Config{
		Name:        "test_agent",
		Description: "A simple agent for testing console runner commands",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				yield(&session.Event{
					LLMResponse: model.LLMResponse{
						Content: genai.NewContentFromText("I am a test agent response", genai.RoleModel),
					},
				}, nil)
			}
		},
	})
	if err != nil {
		outWriter.Close()
		t.Fatalf("failed to create agent: %v", err)
	}

	launcherConfig := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(testAgent),
	}

	subLauncher := NewLauncher().(*consoleLauncher)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run the console loop
	_ = subLauncher.Run(ctx, launcherConfig)

	// Close standard output writer to let io.ReadAll complete
	outWriter.Close()

	return <-outChan
}

func TestConsoleLauncher_Commands(t *testing.T) {
	t.Run("help command prints help details", func(t *testing.T) {
		output := runConsoleWithInput(t, "/help\n/exit\n")
		expected := []string{
			"Available commands:",
			"/help  - Show this help message",
			"/clear - Clear the terminal screen",
			"/exit  - Exit the console session",
			"/quit  - Exit the console session",
		}
		for _, term := range expected {
			if !strings.Contains(output, term) {
				t.Errorf("expected output to contain %q, but got:\n%s", term, output)
			}
		}
	})

	t.Run("clear command prints clearing escape sequence", func(t *testing.T) {
		output := runConsoleWithInput(t, "/clear\n/exit\n")
		clearSeq := "\033[H\033[2J"
		if !strings.Contains(output, clearSeq) {
			t.Errorf("expected output to contain the ANSI clear escape sequence, but got:\n%s", output)
		}
	})

	t.Run("unknown command prints command not found and help hint", func(t *testing.T) {
		output := runConsoleWithInput(t, "/invalidcommand\n/exit\n")
		expected := "Unknown command: /invalidcommand. Type /help for a list of available commands."
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, but got:\n%s", expected, output)
		}
	})

	t.Run("quit exits the console", func(t *testing.T) {
		output := runConsoleWithInput(t, "/quit\n")
		if !strings.Contains(output, "Exiting ADK Console...") {
			t.Errorf("expected output to contain graceful exit message, but got:\n%s", output)
		}
	})

	t.Run("empty input lines are ignored", func(t *testing.T) {
		output := runConsoleWithInput(t, "\n  \n/exit\n")
		// The agent should not be triggered if inputs are empty
		if strings.Contains(output, "I am a test agent response") {
			t.Errorf("expected empty inputs to be ignored, but agent was triggered:\n%s", output)
		}
	})
}
