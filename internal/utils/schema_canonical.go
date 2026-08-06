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

// canonicalizeTo recursively serializes v into buf with sorted map keys.
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
			for k, valItem := range val {
				if isSafeString(k) {
					writeString(buf, k)
				} else {
					keyBytes, err := json.Marshal(k)
					if err != nil {
						return err
					}
					buf.Write(keyBytes)
				}
				buf.WriteByte(':')
				if err := canonicalizeTo(buf, valItem); err != nil {
					return err
				}
			}
			buf.WriteByte('}')
			return nil
		}

		var keys []string
		var localKeys [16]string
		if n <= 16 {
			keys = localKeys[:0]
		} else {
			keys = make([]string, 0, n)
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
			if isSafeString(k) {
				writeString(buf, k)
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
			writeString(buf, val)
		} else {
			b, err := json.Marshal(val)
			if err != nil {
				return err
			}
			buf.Write(b)
		}
		return nil

	case float64:
		var scratch [64]byte
		b := strconv.AppendFloat(scratch[:0], val, 'g', -1, 64)
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

// isSafeString reports whether the string is safe to serialize directly
// inside double quotes without any escaping or HTML injection safety concerns.
func isSafeString(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c >= 0x7f || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			return false
		}
	}
	return true
}

// writeString serializes the safe string s directly wrapped in double quotes.
func writeString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	buf.WriteString(s)
	buf.WriteByte('"')
}
