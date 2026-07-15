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

package console_test

import (
	"testing"

	"google.golang.org/adk/cmd/launcher/console"
)

func TestConsoleLauncher_BasicProperties(t *testing.T) {
	cl := console.NewLauncher()
	if cl == nil {
		t.Fatal("expected non-nil launcher")
	}

	if got, want := cl.Keyword(), "console"; got != want {
		t.Errorf("cl.Keyword() = %q; want %q", got, want)
	}

	if got := cl.SimpleDescription(); got == "" {
		t.Error("cl.SimpleDescription() returned empty string")
	}
}

func TestConsoleLauncher_Parse(t *testing.T) {
	cl := console.NewLauncher()

	// Test valid parsing
	remaining, err := cl.Parse([]string{"--streaming_mode", "sse"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("expected no remaining args, got %v", remaining)
	}

	// Test invalid streaming mode
	clInvalid := console.NewLauncher()
	_, err = clInvalid.Parse([]string{"--streaming_mode", "invalid_mode"})
	if err == nil {
		t.Error("expected error for invalid streaming mode, got nil")
	}
}
