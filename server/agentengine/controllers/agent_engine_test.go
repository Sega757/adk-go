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

package controllers_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"google.golang.org/adk/v2/server/agentengine/controllers"
	"google.golang.org/adk/v2/server/agentengine/controllers/method"
	"google.golang.org/adk/v2/session"
)

type mockMethodHandler struct {
	name   string
	called bool
}

func (m *mockMethodHandler) Name() string { return m.name }

func (m *mockMethodHandler) Handle(ctx context.Context, rw http.ResponseWriter, payload []byte) error {
	m.called = true
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(`{"status":"ok"}`))
	return nil
}

func (m *mockMethodHandler) Metadata() (*structpb.Struct, error) {
	return nil, nil
}

var _ method.MethodHandler = (*mockMethodHandler)(nil)

// TestAgentEngineControllerBodyLimit verifies that request bodies exceeding
// maxPayloadSize are rejected with HTTP 413 StatusRequestEntityTooLarge via MaxBytesReader.
func TestAgentEngineControllerBodyLimit(t *testing.T) {
	t.Parallel()

	const limit = 1024
	handler := &mockMethodHandler{name: "test_method"}
	ctrl, err := controllers.NewAgentEngineAPIController(session.InMemoryService(), 30*time.Second, limit, []method.MethodHandler{handler})
	if err != nil {
		t.Fatalf("NewAgentEngineAPIController error = %v", err)
	}

	tests := []struct {
		name       string
		payload    string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "valid small payload",
			payload:    `{"class_method":"test_method","input":{"param":"small"}}`,
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "oversized payload",
			payload:    fmt.Sprintf(`{"class_method":"test_method","input":{"param":%q}}`, strings.Repeat("x", limit+100)),
			wantStatus: http.StatusRequestEntityTooLarge,
			wantCalled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler.called = false
			req := httptest.NewRequest(http.MethodPost, "/reasoning_engine", bytes.NewBufferString(tc.payload))
			rec := httptest.NewRecorder()

			ctrl.Query(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body = %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if handler.called != tc.wantCalled {
				t.Errorf("handler.called = %v, want %v", handler.called, tc.wantCalled)
			}
		})
	}
}
