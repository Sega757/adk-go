// Copyright 2025 Google LLC
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
	"google.golang.org/genai"

	"google.golang.org/adk/v2/session"
)

// Content is a convenience function that returns the genai.Content
// in the event.
func Content(ev *session.Event) *genai.Content {
	if ev == nil {
		return nil
	}
	return ev.LLMResponse.Content
}

// PartContainer represents types that contain genai.Part items.
type PartContainer interface {
	*genai.Content | []*genai.Part | *session.Event
}

// extractParts extracts the slice of parts from a PartContainer with zero allocations.
func extractParts[T PartContainer](c T) []*genai.Part {
	switch v := any(c).(type) {
	case *genai.Content:
		if v == nil {
			return nil
		}
		return v.Parts
	case []*genai.Part:
		return v
	case *session.Event:
		if v == nil || v.Content == nil {
			return nil
		}
		return v.Content.Parts
	default:
		return nil
	}
}

// HasFunctionCalls reports whether container c contains any FunctionCall.
// It provides a zero-allocation alternative to checking len(FunctionCalls(c)) > 0.
func HasFunctionCalls[T PartContainer](c T) bool {
	for _, p := range extractParts(c) {
		if p != nil && p.FunctionCall != nil {
			return true
		}
	}
	return false
}

// HasFunctionResponses reports whether container c contains any FunctionResponse.
// It provides a zero-allocation alternative to checking len(FunctionResponses(c)) > 0.
func HasFunctionResponses[T PartContainer](c T) bool {
	for _, p := range extractParts(c) {
		if p != nil && p.FunctionResponse != nil {
			return true
		}
	}
	return false
}

// FunctionCalls extracts all FunctionCall parts from the content.
func FunctionCalls(c *genai.Content) (ret []*genai.FunctionCall) {
	if c == nil {
		return nil
	}
	for _, p := range c.Parts {
		if p != nil && p.FunctionCall != nil {
			if ret == nil {
				// Pre-allocate slice capacity upon finding first match to avoid reallocations.
				ret = make([]*genai.FunctionCall, 0, len(c.Parts))
			}
			ret = append(ret, p.FunctionCall)
		}
	}
	return ret
}

// FunctionResponses extracts all FunctionResponse parts from the content.
func FunctionResponses(c *genai.Content) (ret []*genai.FunctionResponse) {
	if c == nil {
		return nil
	}
	for _, p := range c.Parts {
		if p != nil && p.FunctionResponse != nil {
			if ret == nil {
				// Pre-allocate slice capacity upon finding first match to avoid reallocations.
				ret = make([]*genai.FunctionResponse, 0, len(c.Parts))
			}
			ret = append(ret, p.FunctionResponse)
		}
	}
	return ret
}

// TextParts extracts all Text parts from the content.
func TextParts(c *genai.Content) (ret []string) {
	if c == nil {
		return nil
	}
	for _, p := range c.Parts {
		if p != nil && p.Text != "" {
			if ret == nil {
				// Pre-allocate slice capacity upon finding first match to avoid reallocations.
				ret = make([]string, 0, len(c.Parts))
			}
			ret = append(ret, p.Text)
		}
	}
	return ret
}

// IsZeroPart reports whether p is nil or points to a zero-value genai.Part.
// It provides a reflection-free, zero-allocation alternative to reflect.ValueOf(*p).IsZero().
func IsZeroPart(p *genai.Part) bool {
	if p == nil {
		return true
	}
	return p.Text == "" &&
		!p.Thought &&
		len(p.ThoughtSignature) == 0 &&
		p.FunctionCall == nil &&
		p.FunctionResponse == nil &&
		p.InlineData == nil &&
		p.FileData == nil &&
		p.ExecutableCode == nil &&
		p.CodeExecutionResult == nil &&
		p.MediaResolution == nil &&
		p.VideoMetadata == nil &&
		p.ToolCall == nil &&
		p.ToolResponse == nil &&
		len(p.PartMetadata) == 0
}

// FunctionDecls extracts all Function declarations from the GenerateContentConfig.
func FunctionDecls(c *genai.GenerateContentConfig) (ret []*genai.FunctionDeclaration) {
	if c == nil {
		return nil
	}
	for _, t := range c.Tools {
		if t != nil && len(t.FunctionDeclarations) > 0 {
			if ret == nil {
				// Pre-allocate slice capacity upon finding first match to avoid reallocations.
				ret = make([]*genai.FunctionDeclaration, 0, len(c.Tools)*len(t.FunctionDeclarations))
			}
			ret = append(ret, t.FunctionDeclarations...)
		}
	}
	return ret
}
