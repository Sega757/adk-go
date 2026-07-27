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
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

func TestConsoleLauncher_ExitCommand(t *testing.T) {
	// Re-direct Stdin and Stdout
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	rStdin, wStdin, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	defer rStdin.Close()

	rStdout, wStdout, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	defer rStdout.Close()

	os.Stdin = rStdin
	os.Stdout = wStdout

	// Write /exit command to stdin
	go func() {
		_, _ = wStdin.Write([]byte("/exit\n"))
		wStdin.Close()
	}()

	// Capture stdout in a separate goroutine
	stdoutChan := make(chan string, 1)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, rStdout)
		stdoutChan <- buf.String()
	}()

	// Create a dummy agent
	var agentRunCount int32
	greet := workflow.NewFunctionNode("greet",
		func(_ agent.Context, in string) (string, error) {
			atomic.AddInt32(&agentRunCount, 1)
			return "Hello " + in, nil
		},
		workflow.NodeConfig{},
	)
	dummyAgent, err := workflowagent.New(workflowagent.Config{
		Name:  "dummy",
		Edges: workflow.Chain(workflow.Start, greet),
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	config := &launcher.Config{
		SessionService: session.InMemoryService(),
		AgentLoader:    agent.NewSingleLoader(dummyAgent),
	}

	l := NewLauncher()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = l.Run(ctx, config)
	_ = wStdout.Close()

	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	stdout := <-stdoutChan

	// Verify that welcome banner was printed and goodbye was printed
	if !strings.Contains(stdout, "Welcome to the Agent Development Kit CLI!") {
		t.Errorf("expected stdout to contain welcome banner, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Goodbye!") {
		t.Errorf("expected stdout to contain goodbye message, got: %q", stdout)
	}

	// Verify the agent was never run
	if runCount := atomic.LoadInt32(&agentRunCount); runCount != 0 {
		t.Errorf("expected agent run count to be 0, got %d", runCount)
	}
}

func TestConsoleLauncher_QuitCommand(t *testing.T) {
	// Re-direct Stdin and Stdout
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	rStdin, wStdin, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	defer rStdin.Close()

	rStdout, wStdout, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	defer rStdout.Close()

	os.Stdin = rStdin
	os.Stdout = wStdout

	// Write /quit command to stdin
	go func() {
		_, _ = wStdin.Write([]byte("/quit\n"))
		wStdin.Close()
	}()

	// Capture stdout in a separate goroutine
	stdoutChan := make(chan string, 1)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, rStdout)
		stdoutChan <- buf.String()
	}()

	// Create a dummy agent
	greet := workflow.NewFunctionNode("greet",
		func(_ agent.Context, in string) (string, error) {
			return "Hello " + in, nil
		},
		workflow.NodeConfig{},
	)
	dummyAgent, err := workflowagent.New(workflowagent.Config{
		Name:  "dummy",
		Edges: workflow.Chain(workflow.Start, greet),
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	config := &launcher.Config{
		SessionService: session.InMemoryService(),
		AgentLoader:    agent.NewSingleLoader(dummyAgent),
	}

	l := NewLauncher()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = l.Run(ctx, config)
	_ = wStdout.Close()

	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	stdout := <-stdoutChan

	if !strings.Contains(stdout, "Goodbye!") {
		t.Errorf("expected stdout to contain goodbye message, got: %q", stdout)
	}
}

func TestConsoleLauncher_EmptyAndWhitespaceInput(t *testing.T) {
	// Re-direct Stdin and Stdout
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	rStdin, wStdin, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	defer rStdin.Close()

	rStdout, wStdout, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	defer rStdout.Close()

	os.Stdin = rStdin
	os.Stdout = wStdout

	// Write empty strings, whitespaces and then /exit
	go func() {
		_, _ = wStdin.Write([]byte("\n"))
		_, _ = wStdin.Write([]byte("   \n"))
		_, _ = wStdin.Write([]byte("\t\n"))
		_, _ = wStdin.Write([]byte("/exit\n"))
		wStdin.Close()
	}()

	// Capture stdout in a separate goroutine
	stdoutChan := make(chan string, 1)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, rStdout)
		stdoutChan <- buf.String()
	}()

	// Create a dummy agent
	var agentRunCount int32
	greet := workflow.NewFunctionNode("greet",
		func(_ agent.Context, in string) (string, error) {
			atomic.AddInt32(&agentRunCount, 1)
			return "Hello " + in, nil
		},
		workflow.NodeConfig{},
	)
	dummyAgent, err := workflowagent.New(workflowagent.Config{
		Name:  "dummy",
		Edges: workflow.Chain(workflow.Start, greet),
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	config := &launcher.Config{
		SessionService: session.InMemoryService(),
		AgentLoader:    agent.NewSingleLoader(dummyAgent),
	}

	l := NewLauncher()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = l.Run(ctx, config)
	_ = wStdout.Close()

	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	stdout := <-stdoutChan

	// Verify the agent was never run
	if runCount := atomic.LoadInt32(&agentRunCount); runCount != 0 {
		t.Errorf("expected agent run count to be 0, got %d", runCount)
	}

	// Verify that "User -> " prompt is printed for empty/whitespace inputs
	// Expected:
	// Welcome banner...
	// User -> (initially)
	// User -> (empty input 1)
	// User -> (empty input 2)
	// User -> (empty input 3)
	// Goodbye!
	occurrences := strings.Count(stdout, "User ->")
	if occurrences < 4 {
		t.Errorf("expected at least 4 occurrences of 'User ->' prompt, got %d. Output:\n%s", occurrences, stdout)
	}
}
