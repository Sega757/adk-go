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

package sessionutils

import (
	"reflect"
	"testing"
)

func TestExtractStateDeltas(t *testing.T) {
	tests := []struct {
		name                           string
		delta                          map[string]any
		wantApp, wantUser, wantSession map[string]any
	}{
		{
			name:        "nil delta",
			delta:       nil,
			wantApp:     nil,
			wantUser:    nil,
			wantSession: nil,
		},
		{
			name:        "empty delta",
			delta:       map[string]any{},
			wantApp:     nil,
			wantUser:    nil,
			wantSession: nil,
		},
		{
			name: "session state only",
			delta: map[string]any{
				"key1": "value1",
				"key2": 42,
			},
			wantApp:     nil,
			wantUser:    nil,
			wantSession: map[string]any{"key1": "value1", "key2": 42},
		},
		{
			name: "mixed state deltas",
			delta: map[string]any{
				"app:app_setting": "enabled",
				"user:user_pref":  "dark",
				"session_var":     "active",
				"temp:discard_me": "temp_value",
			},
			wantApp:     map[string]any{"app_setting": "enabled"},
			wantUser:    map[string]any{"user_pref": "dark"},
			wantSession: map[string]any{"session_var": "active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotApp, gotUser, gotSession := ExtractStateDeltas(tt.delta)

			if len(gotApp) == 0 && len(tt.wantApp) == 0 {
				// both empty/nil
			} else if !reflect.DeepEqual(gotApp, tt.wantApp) {
				t.Errorf("ExtractStateDeltas() gotApp = %v, want %v", gotApp, tt.wantApp)
			}

			if len(gotUser) == 0 && len(tt.wantUser) == 0 {
				// both empty/nil
			} else if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("ExtractStateDeltas() gotUser = %v, want %v", gotUser, tt.wantUser)
			}

			if len(gotSession) == 0 && len(tt.wantSession) == 0 {
				// both empty/nil
			} else if !reflect.DeepEqual(gotSession, tt.wantSession) {
				t.Errorf("ExtractStateDeltas() gotSession = %v, want %v", gotSession, tt.wantSession)
			}
		})
	}
}

func BenchmarkExtractStateDeltas_Nil(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = ExtractStateDeltas(nil)
	}
}

func BenchmarkExtractStateDeltas_SessionOnly(b *testing.B) {
	delta := map[string]any{
		"key1": "val1",
		"key2": 100,
		"key3": true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ExtractStateDeltas(delta)
	}
}

func BenchmarkExtractStateDeltas_Mixed(b *testing.B) {
	delta := map[string]any{
		"app:key1":  "val1",
		"user:key2": 100,
		"key3":      true,
		"temp:key4": "ignore",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ExtractStateDeltas(delta)
	}
}
