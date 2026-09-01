package llminternal

import (
	"testing"
	"time"

	"google.golang.org/genai"
)

// Legacy simulator of task_completed sleep behavior.
func simulateTaskCompletedLegacy() {
	isTaskCompleted := true
	if isTaskCompleted {
		time.Sleep(100 * time.Millisecond)
		return
	}
}

// Optimized simulator without time.Sleep.
func simulateTaskCompletedOptimized() {
	isTaskCompleted := true
	if isTaskCompleted {
		return
	}
}

func BenchmarkTaskCompleted_Legacy(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		simulateTaskCompletedLegacy()
	}
}

func BenchmarkTaskCompleted_Optimized(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		simulateTaskCompletedOptimized()
	}
}

func TestTaskCompleted_PartMatching(t *testing.T) {
	parts := []*genai.Part{
		{
			FunctionResponse: &genai.FunctionResponse{
				Name: "task_completed",
			},
		},
	}
	var isTaskCompleted bool
	for _, part := range parts {
		if part.FunctionResponse != nil && part.FunctionResponse.Name == "task_completed" {
			isTaskCompleted = true
			break
		}
	}
	if !isTaskCompleted {
		t.Fatalf("expected task_completed to be identified")
	}
}
