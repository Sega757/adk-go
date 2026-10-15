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

package adkrest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/adk/v2/agent"
)

type dummyAgentLoader struct{}

func (d dummyAgentLoader) ListAgents() []string {
	return []string{"app1"}
}

func (d dummyAgentLoader) LoadAgent(name string) (agent.Agent, error) {
	return nil, nil
}

func (d dummyAgentLoader) RootAgent() agent.Agent {
	return nil
}

func TestSecurityHeaders_adkrest(t *testing.T) {
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler := securityHeaders(dummy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status OK, got %d", rr.Code)
	}

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

func TestNormalizePrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"/", ""},
		{"//", ""},
		{"api", "/api"},
		{"/api", "/api"},
		{"api/", "/api"},
		{"/api/", "/api"},
		{"  /v1/custom/  ", "/v1/custom"},
	}

	for _, tt := range tests {
		got := normalizePrefix(tt.input)
		if got != tt.expected {
			t.Errorf("normalizePrefix(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNewServer_URLPrefix(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		reqPath        string
		expectedStatus int
	}{
		{
			name:           "no prefix configured, standard route",
			prefix:         "",
			reqPath:        "/list-apps",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "prefix configured, request with prefix",
			prefix:         "/api",
			reqPath:        "/api/list-apps",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "prefix configured, request without prefix",
			prefix:         "/api",
			reqPath:        "/list-apps",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "prefix without leading slash, request with prefix",
			prefix:         "v1/api/",
			reqPath:        "/v1/api/list-apps",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServer(ServerConfig{
				URLPrefix:   tt.prefix,
				AgentLoader: dummyAgentLoader{},
			})
			if err != nil {
				t.Fatalf("NewServer() error = %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, tt.reqPath, nil)
			rr := httptest.NewRecorder()

			server.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("ServeHTTP(%q) status = %d; want %d", tt.reqPath, rr.Code, tt.expectedStatus)
			}
		})
	}
}
