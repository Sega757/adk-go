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

package agentengine

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/session"
)

func TestSecurityHeaders(t *testing.T) {
	cfg := &launcher.Config{
		SessionService: session.InMemoryService(),
	}

	handler, err := NewHandler(cfg, 10*time.Second, 1024*1024, "test-agent-engine")
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/reasoning_engine", bytes.NewBufferString("{}"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Cache-Control":          "no-store, no-cache, must-revalidate, proxy-revalidate",
		"Pragma":                 "no-cache",
		"Expires":                "0",
	}

	for header, want := range expectedHeaders {
		got := rr.Header().Get(header)
		if got != want {
			t.Errorf("Header %q = %q, want %q", header, got, want)
		}
	}
}

func TestQueryInternalErrorSanitization(t *testing.T) {
	cfg := &launcher.Config{
		SessionService: session.InMemoryService(),
	}

	handler, err := NewHandler(cfg, 10*time.Second, 1024*1024, "test-agent-engine")
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	// Send invalid classMethod to trigger internal handleQuery error
	reqBody := `{"class_method": "invalid_method"}`
	req := httptest.NewRequest(http.MethodPost, "/reasoning_engine", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	wantBody := "internal server error\n"
	if gotBody := rr.Body.String(); gotBody != wantBody {
		t.Errorf("Response body = %q, want %q", gotBody, wantBody)
	}
}
