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

package converters

import (
	"encoding/json"
)

// ToMapStructure converts any to map[string]any.
// We can't use mapstructure library in a way compatible with ADK-python, because genai type fields
// don't have proper field tags.
// TODO(yarolegovich): field annotation PR for genai types.
func ToMapStructure(data any) (map[string]any, error) {
	// Fast-path: return early for nil or empty map inputs to avoid json.Marshal/Unmarshal allocations.
	if data == nil {
		return nil, nil
	}
	if m, ok := data.(map[string]any); ok {
		if m == nil {
			return nil, nil
		}
		if len(m) == 0 {
			return map[string]any{}, nil
		}
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FromMapStructure converts map[string]any to the type parameter T.
// We can't use mapstructure library in a way compatible with ADK-python, because genai type fields
// don't have proper field tags.
// TODO(yarolegovich): field annotation PR for genai types.
func FromMapStructure[T any](data map[string]any) (*T, error) {
	// Fast-path: return zero value struct pointer for nil or empty map inputs to avoid json.Marshal/Unmarshal allocations.
	if len(data) == 0 {
		var zero T
		return &zero, nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var zero T
	if err := json.Unmarshal(bytes, &zero); err != nil {
		return nil, err
	}
	return &zero, nil
}
