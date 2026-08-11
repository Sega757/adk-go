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

// canonicalizeTo recursively serializes v into the provided bytes.Buffer.
// It optimizes performance by avoiding nested buffer allocations, bypassing
// sorting for single-property maps, utilizing a stack-allocated slice buffer
// for map keys (len <= 16), and fast-pathing safe ASCII strings.
func canonicalizeTo(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
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
		if len(val) > 1 {
			sort.Strings(keys)
		}

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			if !writeSafeASCIIString(buf, k) {
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
		if !writeSafeASCIIString(buf, val) {
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

	case nil:
		buf.WriteString("null")
		return nil

	default:
		// Fallback to standard json.Marshal for float64 and other types.
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	}
}

// writeSafeASCIIString checks if s consists entirely of safe ASCII characters
// and writes it directly to buf with enclosing quotes. Returns true if successful.
func writeSafeASCIIString(buf *bytes.Buffer, s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 32 || c >= 127 || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			return false
		}
	}
	buf.WriteByte('"')
	buf.WriteString(s)
	buf.WriteByte('"')
	return true
}
