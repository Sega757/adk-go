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

// isSafeASCII reports whether the string is safe to encode directly as a JSON string
// without any escaping. Characters less than 0x20, non-ASCII characters, double quotes,
// backslashes, and HTML-unsafe characters (<, >, &) must be escaped, so they are not safe.
func isSafeASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c >= 0x7f || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			return false
		}
	}
	return true
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

// canonicalizeTo serializes v into the provided bytes.Buffer with sorted map keys,
// using high-performance fast paths for safe strings, small maps, and primitive types.
func canonicalizeTo(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		n := len(val)
		if n == 0 {
			buf.WriteString("{}")
			return nil
		}
		if n == 1 {
			buf.WriteByte('{')
			for k, item := range val {
				if isSafeASCII(k) {
					buf.WriteByte('"')
					buf.WriteString(k)
					buf.WriteByte('"')
				} else {
					kBytes, err := json.Marshal(k)
					if err != nil {
						return err
					}
					buf.Write(kBytes)
				}
				buf.WriteByte(':')
				if err := canonicalizeTo(buf, item); err != nil {
					return err
				}
			}
			buf.WriteByte('}')
			return nil
		}

		// Use stack array for small maps to avoid heap allocation
		var keys []string
		if n <= 16 {
			var stackKeys [16]string
			keys = stackKeys[:0]
			for k := range val {
				keys = append(keys, k)
			}
		} else {
			keys = make([]string, 0, n)
			for k := range val {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)

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
				kBytes, err := json.Marshal(k)
				if err != nil {
					return err
				}
				buf.Write(kBytes)
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
		if isSafeASCII(val) {
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
