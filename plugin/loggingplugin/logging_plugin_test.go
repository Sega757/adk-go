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

package loggingplugin

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"google.golang.org/genai"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
)

type dummyContext struct {
	agent.Context
}

func (d dummyContext) AgentName() string {
	return "test_agent"
}

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestBeforeModel_SystemInstruction(t *testing.T) {
	tests := []struct {
		name           string
		parts          []*genai.Part
		expectedOutput string
	}{
		{
			name:           "Empty parts",
			parts:          []*genai.Part{},
			expectedOutput: "",
		},
		{
			name: "Single part short",
			parts: []*genai.Part{
				{Text: "You are a helpful assistant."},
			},
			expectedOutput: "System Instruction: 'You are a helpful assistant.'",
		},
		{
			name: "Single part long (>200 chars)",
			parts: []*genai.Part{
				{Text: strings.Repeat("a", 250)},
			},
			expectedOutput: "System Instruction: '" + strings.Repeat("a", 200) + "...'",
		},
		{
			name: "Multi-part short (<200 chars total)",
			parts: []*genai.Part{
				{Text: "You are a helpful assistant. "},
				{Text: "Always be polite."},
			},
			expectedOutput: "System Instruction: 'You are a helpful assistant. Always be polite.'",
		},
		{
			name: "Multi-part long (>200 chars total)",
			parts: []*genai.Part{
				{Text: strings.Repeat("a", 150)},
				{Text: strings.Repeat("b", 100)},
			},
			expectedOutput: "System Instruction: '" + strings.Repeat("a", 150) + strings.Repeat("b", 50) + "...'",
		},
	}

	p := &loggingPlugin{name: "test"}
	ctx := dummyContext{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &model.LLMRequest{
				Config: &genai.GenerateContentConfig{
					SystemInstruction: &genai.Content{
						Parts: tc.parts,
					},
				},
			}

			out := captureOutput(func() {
				_, _ = p.beforeModel(ctx, req)
			})

			if tc.expectedOutput == "" {
				if strings.Contains(out, "System Instruction:") {
					t.Errorf("expected no System Instruction logged, got: %s", out)
				}
			} else {
				if !strings.Contains(out, tc.expectedOutput) {
					t.Errorf("expected output to contain %q, got: %s", tc.expectedOutput, out)
				}
			}
		})
	}
}

func (p *loggingPlugin) formatSystemInstruction(req *model.LLMRequest) string {
	if req.Config != nil && req.Config.SystemInstruction != nil {
		parts := req.Config.SystemInstruction.Parts
		var sysInstruction string
		switch len(parts) {
		case 0:
		case 1:
			text := parts[0].Text
			if len(text) > 200 {
				sysInstruction = text[:200] + "..."
			} else {
				sysInstruction = text
			}
		default:
			var sb strings.Builder
			for _, p := range parts {
				sb.WriteString(p.Text)
				if sb.Len() >= 200 {
					break
				}
			}
			full := sb.String()
			if len(full) > 200 {
				sysInstruction = full[:200] + "..."
			} else {
				sysInstruction = full
			}
		}

		return sysInstruction
	}
	return ""
}

func BenchmarkFormatSystemInstruction(b *testing.B) {
	p := &loggingPlugin{name: "test"}

	b.Run("ZeroParts", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = p.formatSystemInstruction(req)
		}
	})

	b.Run("SinglePartShort", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: "You are a helpful assistant."},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = p.formatSystemInstruction(req)
		}
	})

	b.Run("SinglePartLong", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: strings.Repeat("a", 300)},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = p.formatSystemInstruction(req)
		}
	})

	b.Run("MultiPartShort", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: "You are a helpful assistant. "},
						{Text: "Always be polite."},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = p.formatSystemInstruction(req)
		}
	})

	b.Run("MultiPartLong", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: strings.Repeat("a", 100)},
						{Text: strings.Repeat("b", 100)},
						{Text: strings.Repeat("c", 100)},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = p.formatSystemInstruction(req)
		}
	})
}

func BenchmarkBeforeModel(b *testing.B) {
	p := &loggingPlugin{name: "test"}
	ctx := dummyContext{}

	// Suppress stdout during benchmark
	oldStdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = oldStdout }()

	b.Run("ZeroParts", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.beforeModel(ctx, req)
		}
	})

	b.Run("SinglePartShort", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: "You are a helpful assistant."},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.beforeModel(ctx, req)
		}
	})

	b.Run("SinglePartLong", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: strings.Repeat("a", 300)},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.beforeModel(ctx, req)
		}
	})

	b.Run("MultiPartShort", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: "You are a helpful assistant. "},
						{Text: "Always be polite."},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.beforeModel(ctx, req)
		}
	})

	b.Run("MultiPartLong", func(b *testing.B) {
		req := &model.LLMRequest{
			Config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: strings.Repeat("a", 100)},
						{Text: strings.Repeat("b", 100)},
						{Text: strings.Repeat("c", 100)},
					},
				},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.beforeModel(ctx, req)
		}
	})
}
