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
	"context"
	"strings"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/platform"
	"google.golang.org/adk/v2/session"
)

// TODO: split in proper files/packages.

const afFunctionCallIDPrefix = "adk-"

// PopulateClientFunctionCallID sets the function call ID field if it is empty.
// Since the ID field is optional, some models don't fill the field, but
// the LLMAgent depends on the IDs to map FunctionCall and FunctionResponse events
// in the event stream.
func PopulateClientFunctionCallID(ctx context.Context, c *genai.Content) {
	if c == nil {
		return
	}
	// Direct iteration over Parts avoids allocating intermediate slices via FunctionCalls.
	for _, p := range c.Parts {
		if p != nil && p.FunctionCall != nil && p.FunctionCall.ID == "" {
			p.FunctionCall.ID = GenerateFunctionCallID(ctx)
		}
	}
}

// GenerateFunctionCallID generates a new function call ID. The ID is obtained
// through the platform package, so a UUID provider installed on ctx (see
// platform.WithUUIDProvider) controls it.
func GenerateFunctionCallID(ctx context.Context) string {
	return afFunctionCallIDPrefix + platform.NewUUID(ctx)
}

// RemoveClientFunctionCallID removes the function call ID field that was set
// by populateClientFunctionCallID. This is necessary when FunctionCall or
// FunctionResponse are sent back to the model.
func RemoveClientFunctionCallID(c *genai.Content) {
	if c == nil {
		return
	}
	// Direct single-pass iteration avoids intermediate slice allocations from FunctionCalls/FunctionResponses.
	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		if p.FunctionCall != nil && strings.HasPrefix(p.FunctionCall.ID, afFunctionCallIDPrefix) {
			p.FunctionCall.ID = ""
		}
		if p.FunctionResponse != nil && strings.HasPrefix(p.FunctionResponse.ID, afFunctionCallIDPrefix) {
			p.FunctionResponse.ID = ""
		}
	}
}

// Content is a convenience function that returns the genai.Content
// in the event.
func Content(ev *session.Event) *genai.Content {
	if ev == nil {
		return nil
	}
	return ev.LLMResponse.Content
}

// Belows are useful utilities that help working with genai.Content
// included in types.Event.
// TODO: Use generics.
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

func Must[T agent.Agent](a T, err error) T {
	if err != nil {
		panic(err)
	}
	return a
}

// AppendInstructions appends instructions to the [genai.GenerateContentConfig.SystemInstruction] system instruction.
func AppendInstructions(r *model.LLMRequest, instructions ...string) {
	if len(instructions) == 0 {
		return
	}

	inst := strings.Join(instructions, "\n\n")

	if r.Config == nil {
		r.Config = &genai.GenerateContentConfig{}
	}

	if r.Config.SystemInstruction == nil {
		r.Config.SystemInstruction = genai.NewContentFromText(inst, genai.RoleUser)
		return
	}
	if len(r.Config.SystemInstruction.Parts) > 0 && r.Config.SystemInstruction.Parts[len(r.Config.SystemInstruction.Parts)-1].Text != "" {
		r.Config.SystemInstruction.Parts[len(r.Config.SystemInstruction.Parts)-1].Text += "\n\n" + inst
		return
	}
	r.Config.SystemInstruction.Parts = append(r.Config.SystemInstruction.Parts, genai.NewPartFromText(inst))
}
