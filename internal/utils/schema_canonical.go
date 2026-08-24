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
func CanonicalSchemaJSON(s *jsonschema.Schema) ([]byte, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return canonicalize(v)
}

// canonicalize recursively serializes v with sorted map keys. Slices
// keep their order.
func canonicalize(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := canonicalizeTo(v, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalizeTo writes canonicalized v to the provided buffer to avoid allocation overhead.
func canonicalizeTo(v any, buf *bytes.Buffer) error {
	if v == nil {
		buf.WriteString("null")
		return nil
	}

	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		buf.WriteByte('{')
		if len(val) == 1 {
			// Bypass sorting and slice allocations for single-property maps
			var k string
			var item any
			for k, item = range val {
			}
			if err := writeString(k, buf); err != nil {
				return err
			}
			buf.WriteByte(':')
			if err := canonicalizeTo(item, buf); err != nil {
				return err
			}
		} else if len(val) > 1 {
			// Employ stack-allocated slice buffer for keys to avoid heap allocation
			var localKeys [16]string
			var keys []string
			if len(val) <= 16 {
				keys = localKeys[:0]
			} else {
				keys = make([]string, 0, len(val))
			}

			for k := range val {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for i, k := range keys {
				if i > 0 {
					buf.WriteByte(',')
				}
				if err := writeString(k, buf); err != nil {
					return err
				}
				buf.WriteByte(':')
				if err := canonicalizeTo(val[k], buf); err != nil {
					return err
				}
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

	case string:
		return writeString(val, buf)

	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil

	default:
		// Fallback to standard json.Marshal for numbers and other complex types to ensure correctness
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	}
}

// writeString writes a string directly if it is a safe ASCII string, or falls back to json.Marshal.
func writeString(s string, buf *bytes.Buffer) error {
	if isSafeASCIIString(s) {
		buf.WriteByte('"')
		buf.WriteString(s)
		buf.WriteByte('"')
		return nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	buf.Write(b)
	return nil
}

// isSafeASCIIString checks if the string contains only safe ASCII characters that do not need escaping.
func isSafeASCIIString(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Reject control characters (<0x20), non-ASCII (>=0x7f),
		// double quotes, backslash, and HTML-unsafe characters (<, >, &).
		if c < 0x20 || c >= 0x7f || c == '\\' || c == '"' || c == '<' || c == '>' || c == '&' {
			return false
		}
	}
	return true
}
