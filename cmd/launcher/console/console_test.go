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
	"testing"

	"google.golang.org/adk/agent"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantStreaming agent.StreamingMode
		wantErr       bool
	}{
		{
			name:          "no arguments",
			args:          []string{},
			wantStreaming: "",
			wantErr:       false,
		},
		{
			name:          "streaming mode sse",
			args:          []string{"-streaming_mode", "sse"},
			wantStreaming: agent.StreamingModeSSE,
			wantErr:       false,
		},
		{
			name:          "streaming mode none",
			args:          []string{"-streaming_mode", "none"},
			wantStreaming: agent.StreamingModeNone,
			wantErr:       false,
		},
		{
			name:    "invalid streaming mode",
			args:    []string{"-streaming_mode", "invalid"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLauncher().(*consoleLauncher)
			_, err := l.Parse(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && l.config.streamingMode != tc.wantStreaming {
				t.Errorf("l.config.streamingMode = %q, want %q", l.config.streamingMode, tc.wantStreaming)
			}
		})
	}
}

func TestLauncherMetadata(t *testing.T) {
	l := NewLauncher()
	if l.Keyword() != "console" {
		t.Errorf("l.Keyword() = %q, want %q", l.Keyword(), "console")
	}
	if l.SimpleDescription() == "" {
		t.Error("l.SimpleDescription() should not be empty")
	}
	if l.CommandLineSyntax() == "" {
		t.Error("l.CommandLineSyntax() should not be empty")
	}
}
