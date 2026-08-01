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

package adkrest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/adk/server/adkrest"
	"google.golang.org/adk/session"
)

func TestNewServer_SecurityHeaders(t *testing.T) {
	cfg := adkrest.ServerConfig{
		SessionService: session.InMemoryService(),
	}
	server, err := adkrest.NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/apps/test/users/test/sessions", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	headers := []struct {
		key   string
		value string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate"},
		{"Pragma", "no-cache"},
		{"Expires", "0"},
	}

	for _, h := range headers {
		got := rr.Header().Get(h.key)
		if got != h.value {
			t.Errorf("expected header %q to be %q, got %q", h.key, h.value, got)
		}
	}
}
