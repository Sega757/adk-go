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

package llminternal

import (
	"bytes"
	"fmt"
	"iter"
	"maps"
	"reflect"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
)

// basicRequestProcessor populates the LLMRequest
// with the agent's LLM generation configs.
func basicRequestProcessor(ctx agent.InvocationContext, req *model.LLMRequest, f *Flow) iter.Seq2[*session.Event, error] {
	// reference: adk-python src/google/adk/flows/llm_flows/basic.py
	return func(yield func(*session.Event, error) bool) {
		llmAgent := asLLMAgent(ctx.Agent())
		if llmAgent == nil {
			return // do nothing.
		}

		state := llmAgent.internal()

		req.Config = clone(state.GenerateContentConfig)
		if req.Config == nil {
			req.Config = &genai.GenerateContentConfig{}
		}

		// Set OutputSchema directly if no tools are present or native combo support exists.
		// Otherwise, OutputSchemaRequestProcessor will be used to provide a tool-based workaround.
		//
		// Task-mode agents skip OutputSchema configuration entirely:
		// structured output for tasks is collected via the FinishTaskTool's
		// declaration (the model emits the value inside the finish_task
		// FC args, not as a structured text response).
		if state.Mode != ModeTask && state.OutputSchema != nil && !needOutputSchemaProcessor(state) {
			req.Config.ResponseSchema = state.OutputSchema
			req.Config.ResponseMIMEType = "application/json"
		}

		// TODO: missing features
		//  populate LLMRequest LiveConnectConfig setting
	}
}

// clone returns a deep copy of the src.
// NOTE: this does not work for types with unexported fields.
func clone[M any](src M) M {
	switch v := any(src).(type) {
	case *genai.Content:
		if v == nil {
			var zero M
			return zero
		}
		return any(cloneContent(v)).(M)
	case genai.Content:
		cl := cloneContent(&v)
		if cl == nil {
			var zero M
			return zero
		}
		return any(*cl).(M)
	}

	val := reflect.ValueOf(src)

	// Handle nil pointers
	if val.Kind() == reflect.Pointer && val.IsNil() {
		var zero M
		return zero
	}

	srcIsPointer := val.Kind() == reflect.Pointer

	// Dereference pointer to get the underlying value
	if srcIsPointer {
		val = val.Elem()
	}

	// Create a new instance of the same type
	newVal := reflect.New(val.Type()).Elem()

	// Recursively copy fields
	deepCopy(val, newVal)

	// Return as the original type
	if srcIsPointer {
		return newVal.Addr().Interface().(M)
	}
	return newVal.Interface().(M)
}

// deepCopy copies src to dst using reflect.
// Performance-optimized by Bolt: copies struct fields and slice elements directly
// into their destination memory (in-place) without allocating a temporary Value via reflect.New.
// Also skips deep copying key/value of map entries if they are basic/scalar types.
func deepCopy(src, dst reflect.Value) {
	switch src.Kind() {
	case reflect.Struct:
		t := src.Type()
		for i := 0; i < src.NumField(); i++ {
			if !t.Field(i).IsExported() {
				panic(fmt.Sprintf("deepCopy: unexported field %q in type %q", t.Field(i).Name, t.Name()))
			}
			// Copy directly into the destination field without allocating a temporary copy.
			deepCopy(src.Field(i), dst.Field(i))
		}
	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			// Copy directly into the destination slice element without allocating a temporary copy.
			deepCopy(src.Index(i), dst.Index(i))
		}
	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			var keyCopy reflect.Value
			if isCopyRequired(key.Kind()) {
				keyCopy = reflect.New(key.Type()).Elem()
				deepCopy(key, keyCopy)
			} else {
				keyCopy = key
			}

			val := src.MapIndex(key)
			var valCopy reflect.Value
			if isCopyRequired(val.Kind()) {
				valCopy = reflect.New(val.Type()).Elem()
				deepCopy(val, valCopy)
			} else {
				valCopy = val
			}

			dst.SetMapIndex(keyCopy, valCopy)
		}
	case reflect.Pointer:
		if src.IsNil() {
			return
		}
		// Create a new pointer and deep copy the underlying value
		newPtr := reflect.New(src.Elem().Type())
		deepCopy(src.Elem(), newPtr.Elem())
		dst.Set(newPtr)
	default:
		// For basic types, direct assignment is sufficient
		dst.Set(src)
	}
}

// isCopyRequired returns true if the reflect.Kind might contain pointers
// or reference types that require recursive deep copying to avoid sharing.
func isCopyRequired(k reflect.Kind) bool {
	switch k {
	case reflect.Struct, reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface:
		return true
	default:
		return false
	}
}

// cloneContent creates a deep copy of genai.Content without reflection.
func cloneContent(c *genai.Content) *genai.Content {
	if c == nil {
		return nil
	}
	res := &genai.Content{
		Role: c.Role,
	}
	if len(c.Parts) > 0 {
		res.Parts = make([]*genai.Part, len(c.Parts))
		for i, p := range c.Parts {
			res.Parts[i] = clonePart(p)
		}
	}
	return res
}

// clonePart creates a deep copy of genai.Part without reflection.
func clonePart(p *genai.Part) *genai.Part {
	if p == nil {
		return nil
	}
	res := *p
	if len(p.ThoughtSignature) > 0 {
		res.ThoughtSignature = bytes.Clone(p.ThoughtSignature)
	}
	if p.FunctionCall != nil {
		fc := *p.FunctionCall
		if fc.Args != nil {
			fc.Args = maps.Clone(fc.Args)
		}
		res.FunctionCall = &fc
	}
	if p.FunctionResponse != nil {
		fr := *p.FunctionResponse
		if fr.Response != nil {
			fr.Response = maps.Clone(fr.Response)
		}
		res.FunctionResponse = &fr
	}
	if p.InlineData != nil {
		id := *p.InlineData
		if len(id.Data) > 0 {
			id.Data = bytes.Clone(id.Data)
		}
		res.InlineData = &id
	}
	if p.FileData != nil {
		fd := *p.FileData
		res.FileData = &fd
	}
	if p.ExecutableCode != nil {
		ec := *p.ExecutableCode
		res.ExecutableCode = &ec
	}
	if p.CodeExecutionResult != nil {
		cer := *p.CodeExecutionResult
		res.CodeExecutionResult = &cer
	}
	if p.MediaResolution != nil {
		mr := *p.MediaResolution
		res.MediaResolution = &mr
	}
	if p.VideoMetadata != nil {
		vm := *p.VideoMetadata
		res.VideoMetadata = &vm
	}
	if p.ToolCall != nil {
		tc := *p.ToolCall
		res.ToolCall = &tc
	}
	if p.ToolResponse != nil {
		tr := *p.ToolResponse
		res.ToolResponse = &tr
	}
	if len(p.PartMetadata) > 0 {
		res.PartMetadata = maps.Clone(p.PartMetadata)
	}
	return &res
}
