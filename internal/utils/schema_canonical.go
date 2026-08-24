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
	"slices"

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
	buf.Grow(len(raw))
	if err := canonicalizeTo(v, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalize recursively serializes v with sorted map keys. Slices
// keep their order. Primitive values are encoded via json.Marshal.
// Performance-optimized by Bolt: utilizes canonicalizeTo with a shared bytes.Buffer,
// stack-allocated slice buffers, single-property map sort bypass, and an ASCII string fast-path.
func canonicalize(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := canonicalizeTo(v, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func canonicalizeTo(v any, buf *bytes.Buffer) error {
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			buf.WriteString("null")
			return nil
		}
		buf.WriteByte('{')
		n := len(val)
		if n == 0 {
			buf.WriteByte('}')
			return nil
		}

		if n == 1 {
			// Single-property map: bypass key sorting entirely!
			for k, vItem := range val {
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
				if err := canonicalizeTo(vItem, buf); err != nil {
					return err
				}
			}
			buf.WriteByte('}')
			return nil
		}

		// Use stack-allocated slice buffer if key count <= 16 to avoid heap allocations
		var localKeys [16]string
		var keys []string
		if n <= 16 {
			keys = localKeys[:0]
		} else {
			keys = make([]string, 0, n)
		}

		for k := range val {
			keys = append(keys, k)
		}
		slices.Sort(keys)

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

// isSafeASCII returns true if the string can be safely serialized inside double quotes
// without any backslash escaping or HTML/Unicode safety transformations.
func isSafeASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || c == '\\' || c < 0x20 || c >= 0x80 || c == '<' || c == '>' || c == '&' {
			return false
		}
	}
	return true
}
