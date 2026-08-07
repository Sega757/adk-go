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
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/google/jsonschema-go/jsonschema"
)

// CanonicalSchemaJSON marshals the schema to JSON, parses it back, and
// re-emits it with object keys sorted alphabetically (recursively).
// Arrays are preserved in their original order.
// Optimized by Bolt: uses a shared bytes.Buffer, stack-allocated key buffers,
// and selective fast-paths for safe strings, booleans, and floats.
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
	if err := canonicalizeTo(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// isSafeString checks if the string has only characters that do not need
// any JSON or HTML escaping. This allows us to bypass json.Marshal.
func isSafeString(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		// ASCII characters that don't need escaping in JSON or standard HTML:
		// excluding control characters, double quote, backslash, and <, >, &
		if c < 0x20 || c > 0x7e || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			return false
		}
	}
	return true
}

// canonicalizeTo recursively serializes v directly to the provided buffer
// with sorted map keys, minimizing heap allocations.
func canonicalizeTo(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		var keys []string
		// Use a stack-allocated array for small maps (size <= 16) to avoid heap allocation.
		if len(val) <= 16 {
			var kBuf [16]string
			keys = kBuf[:0]
			for k := range val {
				keys = append(keys, k)
			}
		} else {
			keys = make([]string, 0, len(val))
			for k := range val {
				keys = append(keys, k)
			}
		}
		// Bypass key sorting if map has 1 or fewer properties.
		if len(keys) > 1 {
			sort.Strings(keys)
		}

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			if isSafeString(k) {
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
			if err := canonicalizeTo(buf, val[k]); err != nil {
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
			if err := canonicalizeTo(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil

	case string:
		if isSafeString(val) {
			buf.WriteByte('"')
			buf.WriteString(val)
			buf.WriteByte('"')
		} else {
			b, err := json.Marshal(val)
			if err != nil {
				return err
			}
			buf.Write(b)
		}
		return nil

	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil

	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return fmt.Errorf("json: unsupported value: %f", val)
		}
		var fBuf [64]byte
		b := strconv.AppendFloat(fBuf[:0], val, 'g', -1, 64)
		buf.Write(b)
		return nil

	case nil:
		buf.WriteString("null")
		return nil

	default:
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	}
}
