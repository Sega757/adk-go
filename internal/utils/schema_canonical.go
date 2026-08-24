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
	buf.Grow(len(raw))
	if err := canonicalizeTo(v, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalize recursively serializes v with sorted map keys. Slices
// keep their order. Primitive values are encoded via json.Marshal.
func canonicalize(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := canonicalizeTo(v, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeString encodes a JSON string and writes it to buf.
// Uses a safe fast-path for ASCII-only strings to avoid allocations.
func writeString(buf *bytes.Buffer, s string) {
	safe := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Fall back to json.Marshal if there is a backslash, double quote,
		// HTML unsafe character, control character, or non-ASCII byte.
		if c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' || c < 0x20 || c >= 0x80 {
			safe = false
			break
		}
	}
	if safe {
		buf.WriteByte('"')
		buf.WriteString(s)
		buf.WriteByte('"')
		return
	}

	raw, _ := json.Marshal(s)
	buf.Write(raw)
}

// canonicalizeTo recursively writes the canonical representation of v to buf.
func canonicalizeTo(v any, buf *bytes.Buffer) error {
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
			for k, item := range val {
				writeString(buf, k)
				buf.WriteByte(':')
				if err := canonicalizeTo(item, buf); err != nil {
					return err
				}
			}
			buf.WriteByte('}')
			return nil
		}

		// Stack-allocate a small key slice if possible.
		var keysArr [16]string
		var keys []string
		if len(val) <= 16 {
			keys = keysArr[:0]
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
		switch p := val.(type) {
		case nil:
			buf.WriteString("null")
			return nil
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
		default:
			res, err := json.Marshal(p)
			if err != nil {
				return err
			}
			buf.Write(res)
			return nil
		}
	}
}
