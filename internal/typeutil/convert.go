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
// Optimized by Bolt: uses isJSONSafe to bypass expensive json marshal/unmarshal when type
// is float64, string, bool, or untyped nil and resolvedSchema is nil.
func ConvertToWithJSONSchema[From, To any](v From, resolvedSchema *jsonschema.Resolved) (To, error) {
	var zero To

	if resolvedSchema == nil {
		if isJSONSafe(v) {
			if typed, ok := any(v).(To); ok {
				return typed, nil
			}
			// If types differ (e.g. converting a custom float64/string type),
			// fall back to json marshal/unmarshal to do the proper conversion.
		}
	}

	rawArgs, err := json.Marshal(v)
	if err != nil {
		return zero, err
	}
	if resolvedSchema != nil {
		// See https://github.com/google/jsonschema-go/issues/23: in order to
		// validate, we must validate against a map[string]any. Struct validation
		// does not work as it cannot account for `omitempty` or custom marshalling.
		var m map[string]any
		if err := json.Unmarshal(rawArgs, &m); err != nil {
			return zero, err
		}
		if err := resolvedSchema.Validate(m); err != nil {
			return zero, err
		}
	}
	var typed To
	if err := json.Unmarshal(rawArgs, &typed); err != nil {
		return zero, err
	}
	return typed, nil
}

// isJSONSafe returns true if the value v is of type float64, string, bool,
// or untyped nil. This is used to safely bypass json marshal/unmarshal.
// We avoid ints or other types because JSON float64 vs Go int can cause type mismatches,
// and we also avoid typed nil pointers (which can cause panic or incorrect nil checks
// in certain type assertion contexts), as well as reference types like maps/slices
// to prevent data races or incorrect shared references.
func isJSONSafe(v any) bool {
	if v == nil {
		return true
	}
	// Use reflect or fast type assertions. Type assertions are much faster.
	switch v.(type) {
	case float64, string, bool:
		return true
	default:
		// We explicitly do not support ints because Go's json unmarshaling unmarshals numbers
		// to float64 or int depending on target type. To avoid incorrect type assertions
		// like trying to cast a float64 value to an int target or vice versa, we limit
		// to float64, string, and bool.
		// We also avoid typed nil pointers or other reference types to ensure no data races
		// on shared maps or slices.
		return false
	}
}
