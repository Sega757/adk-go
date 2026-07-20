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
	"testing"
)

func TestConsoleMetadata(t *testing.T) {
	l := NewLauncher()
	if l.Keyword() != "console" {
		t.Errorf("Expected keyword 'console', got %q", l.Keyword())
	}
	if l.SimpleDescription() == "" {
		t.Error("Expected non-empty description")
	}
	if l.CommandLineSyntax() == "" {
		t.Error("Expected non-empty command line syntax")
	}
}

func TestConsoleParse(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantRemaining []string
		wantErr       bool
	}{
		{
			name:          "No arguments",
			args:          []string{},
			wantRemaining: []string{},
			wantErr:       false,
		},
		{
			name:          "Valid streaming_mode none",
			args:          []string{"-streaming_mode", "none"},
			wantRemaining: []string{},
			wantErr:       false,
		},
		{
			name:          "Valid streaming_mode sse",
			args:          []string{"-streaming_mode", "sse"},
			wantRemaining: []string{},
			wantErr:       false,
		},
		{
			name:          "Invalid streaming_mode",
			args:          []string{"-streaming_mode", "invalid"},
			wantRemaining: nil,
			wantErr:       true,
		},
		{
			name:          "Valid timeout",
			args:          []string{"-shutdown-timeout", "5s"},
			wantRemaining: []string{},
			wantErr:       false,
		},
		{
			name:          "Extra remaining args",
			args:          []string{"some-extra-arg"},
			wantRemaining: []string{"some-extra-arg"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLauncher().(*consoleLauncher)
			got, err := l.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.wantRemaining) {
					t.Errorf("Parse() got remaining %v, want %v", got, tt.wantRemaining)
					return
				}
				for i := range got {
					if got[i] != tt.wantRemaining[i] {
						t.Errorf("Parse() got remaining %v, want %v", got, tt.wantRemaining)
					}
				}
			}
		})
	}
}
