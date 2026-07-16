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

func TestConsoleLauncherMetadata(t *testing.T) {
	l := NewLauncher()
	if l.Keyword() != "console" {
		t.Errorf("expected keyword to be 'console', got %q", l.Keyword())
	}
	if l.SimpleDescription() == "" {
		t.Error("expected non-empty simple description")
	}
	if l.CommandLineSyntax() == "" {
		t.Error("expected non-empty command line syntax")
	}
}

func TestConsoleLauncherParse(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantMode agent.StreamingMode
		wantErr  bool
	}{
		{
			name:     "valid none streaming mode",
			args:     []string{"-streaming_mode", "none"},
			wantMode: agent.StreamingModeNone,
			wantErr:  false,
		},
		{
			name:     "valid sse streaming mode",
			args:     []string{"-streaming_mode", "sse"},
			wantMode: agent.StreamingModeSSE,
			wantErr:  false,
		},
		{
			name:    "invalid streaming mode",
			args:    []string{"-streaming_mode", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLauncher().(*consoleLauncher)
			_, err := l.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && l.config.streamingMode != tt.wantMode {
				t.Errorf("Parse() streamingMode = %v, wantMode %v", l.config.streamingMode, tt.wantMode)
			}
		})
	}
}
