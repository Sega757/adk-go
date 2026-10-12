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

func TestMergeStates(t *testing.T) {
	tests := []struct {
		name                         string
		appState, userState, sessMap map[string]any
		want                         map[string]any
	}{
		{
			name:      "all nil/empty",
			appState:  nil,
			userState: nil,
			sessMap:   nil,
			want:      nil,
		},
		{
			name:      "session state only",
			appState:  nil,
			userState: nil,
			sessMap:   map[string]any{"key1": "val1"},
			want:      map[string]any{"key1": "val1"},
		},
		{
			name:      "app state only",
			appState:  map[string]any{"app_key": "app_val"},
			userState: nil,
			sessMap:   nil,
			want:      map[string]any{"app:app_key": "app_val"},
		},
		{
			name:      "user state only",
			appState:  nil,
			userState: map[string]any{"user_key": "user_val"},
			sessMap:   nil,
			want:      map[string]any{"user:user_key": "user_val"},
		},
		{
			name:      "all states merged",
			appState:  map[string]any{"app_key": "app_val"},
			userState: map[string]any{"user_key": "user_val"},
			sessMap:   map[string]any{"sess_key": "sess_val"},
			want:      map[string]any{"app:app_key": "app_val", "user:user_key": "user_val", "sess_key": "sess_val"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeStates(tt.appState, tt.userState, tt.sessMap)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeStates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkMergeStates_Empty(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MergeStates(nil, nil, nil)
	}
}

func BenchmarkMergeStates_SessionOnly(b *testing.B) {
	sessMap := map[string]any{
		"key1": "val1",
		"key2": 100,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MergeStates(nil, nil, sessMap)
	}
}

func BenchmarkMergeStates_AppOnly(b *testing.B) {
	appMap := map[string]any{
		"key1": "val1",
		"key2": 100,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MergeStates(appMap, nil, nil)
	}
}

func BenchmarkMergeStates_Mixed(b *testing.B) {
	appMap := map[string]any{"app_k": "app_v"}
	userMap := map[string]any{"user_k": "user_v"}
	sessMap := map[string]any{"sess_k": "sess_v"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MergeStates(appMap, userMap, sessMap)
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
