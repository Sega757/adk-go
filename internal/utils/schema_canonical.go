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
	"strconv"

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
	var buf bytes.Buffer
	if err := canonicalizeTo(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalize recursively serializes v with sorted map keys. Slices
// keep their order. Primitive values are encoded via json.Marshal.
func canonicalize(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := canonicalizeTo(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalizeTo recursively serializes v with sorted map keys into the provided bytes.Buffer.
// Performance-optimized by Bolt: reduces memory allocations and CPU cycles by passing
// a pre-allocated buffer down, avoiding recursive slice/buffer creation, and using
// a fast-path checker for safe JSON strings.
func canonicalizeTo(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		if len(val) == 0 {
			buf.WriteString("{}")
			return nil
		}
		if len(val) == 1 {
			buf.WriteByte('{')
			for k, valItem := range val {
				writeString(buf, k)
				buf.WriteByte(':')
				if err := canonicalizeTo(buf, valItem); err != nil {
					return err
				}
			}
			buf.WriteByte('}')
			return nil
		}

		// Use a stack-allocated buffer for keys if the map is small (<= 16 keys)
		// to completely avoid heap allocation for key slices.
		var keysBuf [16]string
		var keys []string
		if len(val) <= 16 {
			keys = keysBuf[:0]
		} else {
			keys = make([]string, 0, len(val))
		}
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeString(buf, k)
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

	default:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		switch p := val.(type) {
		case bool:
			if p {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
			return nil
		case string:
			writeString(buf, p)
			return nil
		case float64:
			// Formats float64 directly into a stack-allocated temporary buffer,
			// avoiding heap allocations from json.Marshal or strconv.FormatFloat.
			var tmp [64]byte
			b := strconv.AppendFloat(tmp[:0], p, 'g', -1, 64)
			buf.Write(b)
			return nil
		default:
			// Fallback for other scalar values (like ints, etc.)
			b, err := json.Marshal(val)
			if err != nil {
				return err
			}
			buf.Write(b)
			return nil
		}
	}
}

// isSafeString returns true if the string can be wrapped in double quotes
// and written directly to the JSON output without needing escaping.
// Safe strings contain no control characters (< 0x20), backslashes, quotes, or characters that standard json.Marshal HTML-escapes (<, >, &) or non-ASCII characters (>= 0x80).
func isSafeString(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' || c >= 0x80 {
			return false
		}
	}
	return true
}

// writeString writes a string directly to the buffer as a JSON string literal.
// Uses a fast path for safe strings, falling back to json.Marshal for strings needing escaping.
func writeString(buf *bytes.Buffer, s string) {
	if isSafeString(s) {
		buf.WriteByte('"')
		buf.WriteString(s)
		buf.WriteByte('"')
	} else {
		b, _ := json.Marshal(s)
		buf.Write(b)
	}
}
