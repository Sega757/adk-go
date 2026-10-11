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

package llminternal

import (
	"iter"
	"strconv"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/internal/utils"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

// identityRequestProcessor gives the agent identity from the framework.
func identityRequestProcessor(ctx agent.InvocationContext, req *model.LLMRequest, f *Flow) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		llmAgent := asLLMAgent(ctx.Agent())
		if llmAgent == nil {
			return // do nothing.
		}
		if llmAgent.internal().Mode == ModeSingleTurn {
			return
		}

		// Performance-optimized by Bolt: uses direct string concatenation and strconv.Quote
		// instead of slice allocations ([]string), fmt.Sprintf reflection, and strings.Join.
		name := ctx.Agent().Name()
		desc := ctx.Agent().Description()

		var si string
		if desc == "" {
			si = "You are an agent. Your internal name is " + strconv.Quote(name) + "."
		} else {
			si = "You are an agent. Your internal name is " + strconv.Quote(name) + ". The description about you is " + strconv.Quote(desc) + "."
		}

		utils.AppendInstructions(req, si)
	}
}
