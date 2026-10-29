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
	"strings"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/model"
)

// AppendInstructions appends instructions to the [genai.GenerateContentConfig.SystemInstruction] system instruction.
func AppendInstructions(r *model.LLMRequest, instructions ...string) {
	if len(instructions) == 0 {
		return
	}

	var inst string
	if len(instructions) == 1 {
		inst = instructions[0]
	} else {
		inst = strings.Join(instructions, "\n\n")
	}

	if r.Config == nil {
		r.Config = &genai.GenerateContentConfig{}
	}

	if r.Config.SystemInstruction == nil {
		r.Config.SystemInstruction = &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: inst},
			},
		}
		return
	}

	parts := r.Config.SystemInstruction.Parts
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if lastPart.Text != "" {
			lastPart.Text += "\n\n" + inst
			return
		}
	}

	r.Config.SystemInstruction.Parts = append(r.Config.SystemInstruction.Parts, &genai.Part{Text: inst})
}
