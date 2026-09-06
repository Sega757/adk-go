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
	"maps"
	"strings"
)

const (
	appPrefix  = "app:"
	userPrefix = "user:"
	tempPrefix = "temp:"
)

// ExtractStateDeltas splits a single state delta map into three separate maps
// for app, user, and session states based on key prefixes.
// Temporary keys (starting with TempStatePrefix) are ignored.
//
// Optimization Note: Fast-paths empty/nil state deltas to achieve 0 heap allocations (0 B/op).
// Maps are lazily allocated only when matching keys are found, eliminating redundant map allocations.
func ExtractStateDeltas(delta map[string]any) (
	appStateDelta, userStateDelta, sessionStateDelta map[string]any,
) {
	// Fast-path: return nil maps for empty or nil inputs with zero allocations.
	if len(delta) == 0 {
		return nil, nil, nil
	}

	// Lazily allocate result maps only when matching keys are encountered.
	for key, value := range delta {
		if cleanKey, found := strings.CutPrefix(key, appPrefix); found {
			if appStateDelta == nil {
				appStateDelta = make(map[string]any)
			}
			appStateDelta[cleanKey] = value
		} else if cleanKey, found := strings.CutPrefix(key, userPrefix); found {
			if userStateDelta == nil {
				userStateDelta = make(map[string]any)
			}
			userStateDelta[cleanKey] = value
		} else if !strings.HasPrefix(key, tempPrefix) {
			if sessionStateDelta == nil {
				sessionStateDelta = make(map[string]any)
			}
			sessionStateDelta[key] = value
		}
	}
	return appStateDelta, userStateDelta, sessionStateDelta
}

// MergeStates combines app, user, and session state maps into a single map
// for client-side responses, adding the appropriate prefixes back.
func MergeStates(appState, userState, sessionState map[string]any) map[string]any {
	// Pre-allocate map capacity for efficiency.
	totalSize := len(appState) + len(userState) + len(sessionState)
	if totalSize == 0 {
		return make(map[string]any)
	}
	mergedState := make(map[string]any, totalSize)

	maps.Copy(mergedState, sessionState)

	for key, value := range appState {
		mergedState[appPrefix+key] = value
	}

	for key, value := range userState {
		mergedState[userPrefix+key] = value
	}

	return mergedState
}
