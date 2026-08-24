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

// canonicalSchemaJSON marshals the schema to JSON, parses it back, and
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

// canonicalizeTo serializes v to a shared buffer, dramatically reducing heap allocations.
func canonicalizeTo(v any, buf *bytes.Buffer) error {
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		// Performance Optimization: Utilize a stack-allocated array for up to 16 keys
		// to avoid heap allocation overhead of key slices.
		var keys []string
		var kArr [16]string
		if len(val) <= 16 {
			keys = kArr[:0]
		} else {
			keys = make([]string, 0, len(val))
		}
		for k := range val {
			keys = append(keys, k)
		}
		// Performance Optimization: Skip sorting if there is only 1 key or empty.
		if len(val) > 1 {
			sort.Strings(keys)
		}

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			// Performance Optimization: Fast-path string serialization for safe ASCII keys.
			if err := writeString(k, buf); err != nil {
				return err
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

	case string:
		// Performance Optimization: Fast-path string serialization for safe ASCII strings.
		return writeString(val, buf)

	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil

	case nil:
		buf.WriteString("null")
		return nil

	default:
		// Fallback to standard json.Marshal for float64, ints, and complex types.
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	}
}

// isSafeASCII returns true if the string consists entirely of printable, safe ASCII characters.
// This avoids expensive json.Marshal calls for typical schema strings.
func isSafeASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7E || c == '\\' || c == '"' {
			return false
		}
	}
	return true
}

// writeString writes the string with proper quotes, using a fast-path for safe ASCII strings.
func writeString(s string, buf *bytes.Buffer) error {
	if isSafeASCII(s) {
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
