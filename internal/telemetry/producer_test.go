package telemetry

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestProducer_EmitAndWrite(t *testing.T) {
	testFile := "/tmp/test_telemetry.jsonl"
	defer os.Remove(testFile)

	p, err := InitProducer(testFile)
	if err != nil {
		t.Fatalf("Failed to init producer: %v", err)
	}

	event := EventBusPayload{
		EventID: "test-req-001",
		Type:    "STATE_TRANSITION",
		Source:  "ADK_TEST",
		Payload: map[string]interface{}{"status": "ok"},
	}

	p.Emit(event)

	// Ждем флаша горутины
	time.Sleep(100 * time.Millisecond)

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	var readEvent EventBusPayload
	if err := json.Unmarshal(data, &readEvent); err != nil {
		t.Fatalf("Corrupted JSON written: %v", err)
	}

	if readEvent.EventID != event.EventID {
		t.Errorf("Expected EventID %s, got %s", event.EventID, readEvent.EventID)
	}
}
