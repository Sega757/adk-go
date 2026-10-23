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

package preloadmemorytool

import (
	"testing"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/memory"
)

func BenchmarkFormatMemories(b *testing.B) {
	memories := []memory.Entry{
		{
			Author:    "user",
			Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			Content:   genai.NewContentFromText("Hello world, this is a test query for memory retrieval.", genai.RoleUser),
		},
		{
			Author:    "assistant",
			Timestamp: time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
			Content:   genai.NewContentFromText("I can help you with that memory retrieval request.", genai.RoleModel),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = formatMemories(memories)
	}
}

func BenchmarkExtractText_SinglePart(b *testing.B) {
	mem := memory.Entry{
		Content: genai.NewContentFromText("Single part memory content text string.", genai.RoleUser),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = extractText(mem)
	}
}
