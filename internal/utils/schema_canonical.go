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

package utils

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/google/jsonschema-go/jsonschema"
)

// CanonicalSchemaJSON marshals the schema to JSON, parses it back, and
// re-emits it with object keys sorted alphabetically (recursively).
// Arrays are preserved in their original order.
// Performance-optimized by Bolt: uses direct buffer writing,
// key sorting bypass for single-property maps, stack-allocated key slices
// for maps up to size 16, and fast-paths safe ASCII strings.
func CanonicalSchemaJSON(s *jsonschema.Schema) ([]byte, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := canonicalizeTo(v, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalize recursively serializes v with sorted map keys. Slices
// keep their order. Primitive values are encoded via json.Marshal.
// Backwards compatible wrapper around canonicalizeTo.
func canonicalize(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := canonicalizeTo(v, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalizeTo recursively writes the canonicalized representation of v into buf.
func canonicalizeTo(v any, buf *bytes.Buffer) error {
	switch val := v.(type) {
	case nil:
		buf.WriteString("null")
		return nil

	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil

	case string:
		if isSafeASCII(val) {
			buf.WriteByte('"')
			buf.WriteString(val)
			buf.WriteByte('"')
			return nil
		}
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil

	case float64:
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil

	case map[string]any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		var keys []string
		var arr [16]string
		if len(val) <= 16 {
			keys = arr[:0]
		} else {
			keys = make([]string, 0, len(val))
		}
		for k := range val {
			keys = append(keys, k)
		}
		if len(keys) > 1 {
			sort.Strings(keys)
		}

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			if isSafeASCII(k) {
				buf.WriteByte('"')
				buf.WriteString(k)
				buf.WriteByte('"')
			} else {
				keyBytes, err := json.Marshal(k)
				if err != nil {
					return err
				}
				buf.Write(keyBytes)
			}
			buf.WriteByte(':')
			if err := canonicalizeTo(val[k], buf); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
		return nil

	case []any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := canonicalizeTo(item, buf); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil

	default:
		// Safe fallback for other types
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	}
}

// isSafeASCII reports whether string contains only printable ASCII
// characters and does not need any JSON/HTML escaping.
func isSafeASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 32 || b >= 127 || b == '"' || b == '\\' || b == '<' || b == '>' || b == '&' {
			return false
		}
	}
	return true
}
