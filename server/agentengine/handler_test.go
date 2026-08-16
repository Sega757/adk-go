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

	handler, err := NewHandler(cfg, 10*time.Second, 1024*1024, "test-agent-engine-id")
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/reasoning_engine", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	expectedHeaders := map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range expectedHeaders {
		if got := rec.Header().Get(header); got != expectedValue {
			t.Errorf("Header %s = %q; want %q", header, got, expectedValue)
		}
	}
}
