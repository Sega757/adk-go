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

	"google.golang.org/adk/v2/platform"
)

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
	if c == nil || (!HasFunctionCalls(c) && !HasFunctionResponses(c)) {
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
