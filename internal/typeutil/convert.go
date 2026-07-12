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

// Package typeutil is a collection of type handling utility functions.
package typeutil

import (
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
)

// ConvertToWithJSONSchema converts the given value to another type using json marshal/unmarshal.
// If non-nil resolvedSchema is provided, validation against the resolvedSchema will run
// during the conversion.
func ConvertToWithJSONSchema[From, To any](v From, resolvedSchema *jsonschema.Resolved) (To, error) {
	var zero To
	var rawArgs []byte
	var err error
	var validatedMap map[string]any

	if resolvedSchema != nil {
		// Optimization: If the input is already a map[string]any containing only standard JSON-unmarshaled types,
		// we can validate it directly without performing an expensive json.Marshal and json.Unmarshal cycle first.
		if m, ok := any(v).(map[string]any); ok && isJSONSafe(m) {
			if err := resolvedSchema.Validate(m); err != nil {
				return zero, err
			}
			// Do not set validatedMap to m here to avoid returning a shared reference
			// to the input map v, which could cause concurrent data races or mutations.
		} else {
			// See https://github.com/google/jsonschema-go/issues/23: in order to
			// validate, we must validate against a map[string]any. Struct validation
			// does not work as it cannot account for `omitempty` or custom marshalling.
			rawArgs, err = json.Marshal(v)
			if err != nil {
				return zero, err
			}
			if err := json.Unmarshal(rawArgs, &validatedMap); err != nil {
				return zero, err
			}
			if err := resolvedSchema.Validate(validatedMap); err != nil {
				return zero, err
			}
		}
	}

	var typed To
	// If To is map[string]any, we can directly assign and return the validated map
	// (only if it was newly unmarshaled, i.e., not a shared reference)
	// or unmarshal rawArgs if validatedMap is not available.
	if m, ok := any(&typed).(*map[string]any); ok {
		if validatedMap != nil {
			*m = validatedMap
			return typed, nil
		}
		if len(rawArgs) == 0 {
			rawArgs, err = json.Marshal(v)
			if err != nil {
				return zero, err
			}
		}
		if err := json.Unmarshal(rawArgs, m); err != nil {
			return zero, err
		}
		return typed, nil
	}

	if len(rawArgs) == 0 {
		rawArgs, err = json.Marshal(v)
		if err != nil {
			return zero, err
		}
	}
	if err := json.Unmarshal(rawArgs, &typed); err != nil {
		return zero, err
	}
	return typed, nil
}

// isJSONSafe checks if a value consists only of basic JSON-unmarshaled types.
// Numbers must be float64 to align with standard encoding/json.Unmarshal behavior,
// avoiding downstream type-assertion panics (e.g. converting float64 to int).
func isJSONSafe(v any) bool {
	switch val := v.(type) {
	case nil, string, bool, float64:
		return true
	case []any:
		for _, item := range val {
			if !isJSONSafe(item) {
				return false
			}
		}
		return true
	case map[string]any:
		for _, item := range val {
			if !isJSONSafe(item) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
