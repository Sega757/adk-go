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

// Package console provides a simple way to interact with an agent from console application.
package console

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/internal/telemetry"
	"google.golang.org/adk/v2/cmd/launcher/universal"
	"google.golang.org/adk/v2/internal/cli/util"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
)

// consoleConfig contains command-line params for console launcher
type consoleConfig struct {
	streamingMode       agent.StreamingMode
	streamingModeString string // command-line param to be converted to agent.StreamingMode
	otelToCloud         bool
	shutdownTimeout     time.Duration
}

// consoleLauncher allows to interact with an agent in console
type consoleLauncher struct {
	flags  *flag.FlagSet  // flags are used to parse command-line arguments
	config *consoleConfig // config contains parsed command-line parameters
}

// NewLauncher creates new console launcher
func NewLauncher() launcher.SubLauncher {
	config := &consoleConfig{}

	fs := flag.NewFlagSet("console", flag.ContinueOnError)
	fs.StringVar(&config.streamingModeString, "streaming_mode", "",
		fmt.Sprintf("defines streaming mode (%s|%s)", agent.StreamingModeNone, agent.StreamingModeSSE))
	fs.DurationVar(&config.shutdownTimeout, "shutdown-timeout", 2*time.Second, "Console shutdown timeout (i.e. '10s', '2m' - see time.ParseDuration for details) - for waiting for active requests to finish during shutdown")
	fs.BoolVar(&config.otelToCloud, "otel_to_cloud", false, "Enables/disables OpenTelemetry export to GCP: telemetry.googleapis.com. See adk-go/telemetry package for details about supported options, credentials and environment variables.")
	return &consoleLauncher{config: config, flags: fs}
}

// Run implements launcher.SubLauncher. It starts the console interaction loop.
func (l *consoleLauncher) Run(ctx context.Context, config *launcher.Config) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	telemetry, err := telemetry.InitAndSetGlobalOtelProviders(ctx, config, l.config.otelToCloud)
	if err != nil {
		return fmt.Errorf("telemetry initialization failed: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), l.config.shutdownTimeout)
		defer cancel()
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			log.Printf("telemetry shutdown failed: %v", err)
		}
	}()

	// userID and appName are not important at this moment, we can just use any
	userID, appName := "console_user", "console_app"

	sessionService := config.SessionService
	if sessionService == nil {
		sessionService = session.InMemoryService()
	}

	resp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: appName,
		UserID:  userID,
	})
	if err != nil {
		return fmt.Errorf("failed to create the session service: %v", err)
	}

	rootAgent := config.AgentLoader.RootAgent()

	sess := resp.Session

	r, err := runner.New(runner.Config{
		AppName:         appName,
		Agent:           rootAgent,
		SessionService:  sessionService,
		ArtifactService: config.ArtifactService,
		PluginConfig:    config.PluginConfig,
		MemoryService:   config.MemoryService,
	})
	if err != nil {
		return fmt.Errorf("failed to create runner: %v", err)
	}

	inputChan := make(chan string)
	readErrChan := make(chan error, 1)

	go func() {
		defer close(inputChan)
		reader := bufio.NewReader(os.Stdin)
		for {
			userInput, err := reader.ReadString('\n')
			if err != nil {
				readErrChan <- err
				return
			}
			inputChan <- userInput
		}
	}()
	// Print an initial newline to work around PTY/exec buffering issues in some environments.
	fmt.Println()

	// Show welcome banner instructing users how to chat and exit.
	if isTerminal() {
		fmt.Println("\033[1;36m========================================================\033[0m")
		fmt.Println("\033[1;32m  Welcome to ADK Console! Let's chat with your agent.  \033[0m")
		fmt.Println("\033[1;33m  Type your message and press Enter. To exit, press Ctrl+C.\033[0m")
		fmt.Println("\033[1;36m========================================================\033[0m")
	} else {
		fmt.Println("========================================================")
		fmt.Println("  Welcome to ADK Console! Let's chat with your agent.")
		fmt.Println("  Type your message and press Enter. To exit, press Ctrl+C.")
		fmt.Println("========================================================")
	}

	// Resolve "auto" streaming mode once per session (stdout TTY-ness doesn't change).
	defaultStreamingMode := l.config.streamingMode
	if defaultStreamingMode == "" {
		// Stdlib-only terminal heuristic: stdout is a character device.
		// Avoids adding golang.org/x/term dependency (golangci-lint failed to load its export data in CI).
		if isTerminal() {
			defaultStreamingMode = agent.StreamingModeSSE
		} else {
			defaultStreamingMode = agent.StreamingModeNone
		}
	}

	printUserPrompt(isTerminal())

	// pendingInterrupts carries human-input prompts the agent
	// emitted on the previous turn. While non-empty, the next
	// stdin read is interpreted as the answer to its head; once
	// every prompt has an answer the assembled FunctionResponse
	// content is sent as the next "user message" turn.
	var pendingInterrupts []pendingInterrupt
	var pendingResponses []*genai.Part

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-readErrChan:
			if errors.Is(err, io.EOF) {
				fmt.Println("\nEOF detected, exiting...")
				return nil
			}
			log.Fatal(err)
		case userInput, ok := <-inputChan:
			if !ok {
				return nil
			}
			// Drop the line terminator the reader keeps, so the message
			// matches what the web UI submits (no trailing newline).
			userInput = strings.TrimRight(userInput, "\r\n")

			if len(pendingInterrupts) == 0 {
				trimmed := strings.TrimSpace(userInput)
				if trimmed == "" {
					printUserPrompt(isTerminal())
					continue
				}
				if strings.HasPrefix(trimmed, "/") {
					exit, handled := handleSlashCommand(os.Stdout, trimmed, isTerminal())
					if handled {
						if exit {
							return nil
						}
						continue
					}
				}
			}

			var userMsg *genai.Content
			if len(pendingInterrupts) > 0 {
				// Answer the head of the queue; loop back if more
				// prompts remain.
				current := pendingInterrupts[0]
				pendingInterrupts = pendingInterrupts[1:]
				pendingResponses = append(pendingResponses, buildInterruptResponse(current, userInput))
				if len(pendingInterrupts) > 0 {
					renderInterruptPrompt(pendingInterrupts[0], isTerminal())
					continue
				}
				// All answers collected. Bundle every
				// FunctionResponse into one user Content; the
				// workflow runtime routes each to its waiting
				// node by FunctionResponse.ID.
				// TODO: legacy non-workflow agents still pick
				// one agent from the first FunctionResponse.ID
				// and drop the rest.
				userMsg = &genai.Content{
					Role:  string(genai.RoleUser),
					Parts: pendingResponses,
				}
				pendingResponses = nil
			} else {
				userMsg = genai.NewContentFromText(userInput, genai.RoleUser)
			}

			streamingMode := l.config.streamingMode
			if streamingMode == "" {
				streamingMode = defaultStreamingMode
			}

			printAgentPrompt(isTerminal())
			prevText := ""
			printedContent := false
			var finalOutput any
			var collectedEvents []*session.Event
			for event, err := range r.Run(ctx, userID, sess.ID(), userMsg, agent.RunConfig{
				StreamingMode: streamingMode,
			}) {
				if err != nil {
					printErrorPrompt(isTerminal(), err)
				} else {
					collectedEvents = append(collectedEvents, event)
					if event.LLMResponse.Content == nil {
						// Function/terminal nodes carry their result in
						// Event.Output, not model content; keep the latest
						// so a content-less turn still surfaces a result.
						if event.Output != nil {
							finalOutput = event.Output
						}
						continue
					}

					text := ""
					for _, p := range event.LLMResponse.Content.Parts {
						text += p.Text
					}
					if text != "" {
						printedContent = true
					}

					if streamingMode != agent.StreamingModeSSE {
						fmt.Print(text)
						continue
					}

					// In SSE mode, always print partial responses and capture them.
					if !event.IsFinalResponse() {
						fmt.Print(text)
						prevText += text
						continue
					}

					// Only print final response if it doesn't match previously captured text.
					if text != prevText {
						fmt.Print(text)
					}

					prevText = ""
				}
			}

			// If the turn paused on any long-running interrupts,
			// render the first prompt; the next stdin read will
			// be its answer.
			pendingInterrupts = collectPendingInterrupts(collectedEvents)
			if len(pendingInterrupts) > 0 {
				fmt.Println()
				renderInterruptPrompt(pendingInterrupts[0], isTerminal())
				continue
			}
			// A workflow whose terminal node returns a value rather than
			// model content streams no text; surface that result so the
			// turn isn't silent.
			if !printedContent && finalOutput != nil {
				fmt.Print(renderOutput(finalOutput))
			}
			printUserPrompt(isTerminal())
		}
	}
}

func printUserPromptToWriter(w io.Writer, tty bool) {
	if tty {
		fmt.Fprint(w, "\n\033[1;32mUser 👤 ->\033[0m ")
	} else {
		fmt.Fprint(w, "\nUser -> ")
	}
}

func printUserPrompt(tty bool) {
	printUserPromptToWriter(os.Stdout, tty)
}

// handleSlashCommand processes interactive console slash commands (/help, /clear, /exit, /quit).
// It returns exit=true if the console should terminate, and handled=true if the command was recognized
// and processed.
func handleSlashCommand(w io.Writer, command string, tty bool) (exit bool, handled bool) {
	switch command {
	case "/exit", "/quit":
		if tty {
			fmt.Fprintln(w, "\033[1;33mExiting ADK Console...\033[0m")
		} else {
			fmt.Fprintln(w, "Exiting ADK Console...")
		}
		return true, true

	case "/clear":
		fmt.Fprint(w, "\033[H\033[2J")
		printUserPromptToWriter(w, tty)
		return false, true

	case "/help":
		if tty {
			fmt.Fprintln(w, "\n\033[1;36mAvailable commands:\033[0m")
			fmt.Fprintln(w, "  \033[1;32m/help\033[0m  - Show this help message")
			fmt.Fprintln(w, "  \033[1;32m/clear\033[0m - Clear the terminal screen")
			fmt.Fprintln(w, "  \033[1;32m/exit\033[0m  - Exit the console session")
			fmt.Fprintln(w, "  \033[1;32m/quit\033[0m  - Exit the console session")
		} else {
			fmt.Fprintln(w, "\nAvailable commands:")
			fmt.Fprintln(w, "  /help  - Show this help message")
			fmt.Fprintln(w, "  /clear - Clear the terminal screen")
			fmt.Fprintln(w, "  /exit  - Exit the console session")
			fmt.Fprintln(w, "  /quit  - Exit the console session")
		}
		printUserPromptToWriter(w, tty)
		return false, true

	default:
		if tty {
			fmt.Fprintf(w, "\033[1;31mUnknown command: %s. Type /help for a list of available commands.\033[0m\n", command)
		} else {
			fmt.Fprintf(w, "Unknown command: %s. Type /help for a list of available commands.\n", command)
		}
		printUserPromptToWriter(w, tty)
		return false, true
	}
}

func printAgentPrompt(tty bool) {
	if tty {
		fmt.Print("\n\033[1;36mAgent 🤖 ->\033[0m ")
	} else {
		fmt.Print("\nAgent -> ")
	}
}

func printErrorPrompt(tty bool, err error) {
	if tty {
		fmt.Printf("\n\033[1;31mError ❌ -> %v\033[0m\n", err)
	} else {
		fmt.Printf("\nAGENT_ERROR: %v\n", err)
	}
}

// isTerminal returns true if stdout is a character device (TTY).
func isTerminal() bool {
	if fi, err := os.Stdout.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

// renderOutput formats a node's Output value for the console: strings
// as-is, anything else as compact JSON.
func renderOutput(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if b, err := json.Marshal(v); err == nil {
		return string(b)
	}
	return fmt.Sprint(v)
}

// Parse implements launcher.SubLauncher. After parsing console-specific
// arguments returns remaining un-parsed arguments
func (l *consoleLauncher) Parse(args []string) ([]string, error) {
	err := l.flags.Parse(args)
	if err != nil || !l.flags.Parsed() {
		return nil, fmt.Errorf("failed to parse flags: %v", err)
	}
	if l.config.streamingModeString != "" &&
		l.config.streamingModeString != string(agent.StreamingModeNone) &&
		l.config.streamingModeString != string(agent.StreamingModeSSE) {
		return nil, fmt.Errorf("invalid streaming_mode: %v. Should be (%s|%s)", l.config.streamingModeString,
			agent.StreamingModeNone, agent.StreamingModeSSE)
	}
	l.config.streamingMode = agent.StreamingMode(l.config.streamingModeString)
	return l.flags.Args(), nil
}

// Keyword implements launcher.SubLauncher. Returns the command-line keyword for this launcher.
func (l *consoleLauncher) Keyword() string {
	return "console"
}

// CommandLineSyntax implements launcher.SubLauncher. Returns the command-line syntax for the console launcher.
func (l *consoleLauncher) CommandLineSyntax() string {
	return util.FormatFlagUsage(l.flags)
}

// SimpleDescription implements launcher.SubLauncher. Returns a simple description of the console launcher.
func (l *consoleLauncher) SimpleDescription() string {
	return "runs an agent in console mode."
}

// Execute implements launcher.Launcher. It parses arguments and runs the launcher.
func (l *consoleLauncher) Execute(ctx context.Context, config *launcher.Config, args []string) error {
	remainingArgs, err := l.Parse(args)
	if err != nil {
		return fmt.Errorf("cannot parse args: %w", err)
	}
	// do not accept additional arguments
	err = universal.ErrorOnUnparsedArgs(remainingArgs)
	if err != nil {
		return fmt.Errorf("cannot parse all the arguments: %w", err)
	}
	return l.Run(ctx, config)
}
