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

package session

import (
	"strconv"
	"testing"
)

func BenchmarkTrimTempDeltaState_NoTempKeys(b *testing.B) {
	delta := make(map[string]any, 10)
	for i := range 10 {
		delta["key"+strconv.Itoa(i)] = i
	}
	event := &Event{
		Actions: EventActions{
			StateDelta: delta,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = trimTempDeltaState(event)
	}
}

func BenchmarkTrimTempDeltaState_WithTempKeys(b *testing.B) {
	delta := make(map[string]any, 10)
	for i := range 9 {
		delta["key"+strconv.Itoa(i)] = i
	}
	delta[KeyPrefixTemp+"key"] = "temp_val"
	event := &Event{
		Actions: EventActions{
			StateDelta: delta,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = trimTempDeltaState(event)
	}
}
