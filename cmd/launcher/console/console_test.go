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

package console

import (
	"strings"
	"testing"

	"google.golang.org/adk/agent"
)

func TestConsoleLauncher_ParseAndMetadata(t *testing.T) {
	l := NewLauncher()
	cl, ok := l.(*consoleLauncher)
	if !ok {
		t.Fatalf("NewLauncher() did not return *consoleLauncher")
	}

	if l.Keyword() != "console" {
		t.Errorf("expected keyword to be 'console', got %q", l.Keyword())
	}

	if !strings.Contains(l.SimpleDescription(), "console mode") {
		t.Errorf("unexpected SimpleDescription: %q", l.SimpleDescription())
	}

	if l.CommandLineSyntax() == "" {
		t.Errorf("CommandLineSyntax returned empty string")
	}

	// Test Parse with no flags/arguments
	args, err := cl.Parse([]string{})
	if err != nil {
		t.Errorf("Parse with empty args failed: %v", err)
	}
	if len(args) != 0 {
		t.Errorf("expected no remaining args, got %v", args)
	}
	if cl.config.streamingMode != "" {
		t.Errorf("expected empty streamingMode, got %v", cl.config.streamingMode)
	}

	// Test Parse with valid streaming mode sse
	l2 := NewLauncher()
	cl2 := l2.(*consoleLauncher)
	args2, err2 := cl2.Parse([]string{"-streaming_mode", "sse"})
	if err2 != nil {
		t.Errorf("Parse with valid streaming mode sse failed: %v", err2)
	}
	if len(args2) != 0 {
		t.Errorf("expected no remaining args, got %v", args2)
	}
	if cl2.config.streamingMode != agent.StreamingModeSSE {
		t.Errorf("expected streamingMode to be 'sse', got %v", cl2.config.streamingMode)
	}

	// Test Parse with valid streaming mode none
	l3 := NewLauncher()
	cl3 := l3.(*consoleLauncher)
	args3, err3 := cl3.Parse([]string{"-streaming_mode", "none"})
	if err3 != nil {
		t.Errorf("Parse with valid streaming mode none failed: %v", err3)
	}
	if len(args3) != 0 {
		t.Errorf("expected no remaining args, got %v", args3)
	}
	if cl3.config.streamingMode != agent.StreamingModeNone {
		t.Errorf("expected streamingMode to be 'none', got %v", cl3.config.streamingMode)
	}

	// Test Parse with invalid streaming mode
	l4 := NewLauncher()
	cl4 := l4.(*consoleLauncher)
	_, err4 := cl4.Parse([]string{"-streaming_mode", "invalid"})
	if err4 == nil {
		t.Errorf("expected Parse with invalid streaming mode to return error, got nil")
	}
}
