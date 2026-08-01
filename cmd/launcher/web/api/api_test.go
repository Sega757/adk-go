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

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCleanAddress(t *testing.T) {
	tests := []struct {
		input          string
		expectedScheme string
		expectedHost   string
	}{
		{"localhost:8080", "", "localhost:8080"},
		{"http://localhost:8080", "http", "localhost:8080"},
		{"https://localhost:8080", "https", "localhost:8080"},
		{"  http://localhost:8080/  ", "http", "localhost:8080"},
		{"my-frontend-app.com", "", "my-frontend-app.com"},
		{"https://my-frontend-app.com/api", "https", "my-frontend-app.com"},
		{"", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			gotScheme, gotHost := cleanAddress(tc.input)
			if gotScheme != tc.expectedScheme || gotHost != tc.expectedHost {
				t.Errorf("cleanAddress(%q) = (%q, %q), expected (%q, %q)", tc.input, gotScheme, gotHost, tc.expectedScheme, tc.expectedHost)
			}
		})
	}
}

func TestCorsWithArgs(t *testing.T) {
	// A simple dummy handler to wrap
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	tests := []struct {
		name            string
		frontendAddress string
		originHeader    string
		method          string
		expectedStatus  int
		expectedAllow   string
		expectedVary    string
	}{
		{
			name:            "Matching origin without scheme in config",
			frontendAddress: "localhost:8080",
			originHeader:    "http://localhost:8080",
			method:          http.MethodGet,
			expectedStatus:  http.StatusOK,
			expectedAllow:   "http://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "Matching origin with HTTPS scheme in config",
			frontendAddress: "https://my-frontend-app.com",
			originHeader:    "https://my-frontend-app.com",
			method:          http.MethodPost,
			expectedStatus:  http.StatusOK,
			expectedAllow:   "https://my-frontend-app.com",
			expectedVary:    "Origin",
		},
		{
			name:            "No origin header (direct client)",
			frontendAddress: "localhost:8080",
			originHeader:    "",
			method:          http.MethodGet,
			expectedStatus:  http.StatusOK,
			expectedAllow:   "",
			expectedVary:    "",
		},
		{
			name:            "Unmatched/malicious origin",
			frontendAddress: "localhost:8080",
			originHeader:    "http://attacker.com",
			method:          http.MethodGet,
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    "",
		},
		{
			name:            "Matching origin with OPTIONS preflight",
			frontendAddress: "localhost:8080",
			originHeader:    "http://localhost:8080",
			method:          http.MethodOptions,
			expectedStatus:  http.StatusOK,
			expectedAllow:   "http://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "Mismatched origin with OPTIONS preflight",
			frontendAddress: "localhost:8080",
			originHeader:    "http://attacker.com",
			method:          http.MethodOptions,
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    "",
		},
		{
			name:            "Invalid origin header format",
			frontendAddress: "localhost:8080",
			originHeader:    "://invalid-url",
			method:          http.MethodGet,
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    "",
		},
		{
			name:            "Cross-scheme reject: HTTP origin when HTTPS config is set",
			frontendAddress: "https://my-frontend-app.com",
			originHeader:    "http://my-frontend-app.com",
			method:          http.MethodGet,
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    "",
		},
		{
			name:            "Cross-scheme reject: HTTPS origin when HTTP config is set",
			frontendAddress: "http://my-frontend-app.com",
			originHeader:    "https://my-frontend-app.com",
			method:          http.MethodGet,
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			middleware := corsWithArgs(tc.frontendAddress)(dummyHandler)

			req := httptest.NewRequest(tc.method, "/test", nil)
			if tc.originHeader != "" {
				req.Header.Set("Origin", tc.originHeader)
			}

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			gotAllow := rr.Header().Get("Access-Control-Allow-Origin")
			if gotAllow != tc.expectedAllow {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tc.expectedAllow, gotAllow)
			}

			gotVary := rr.Header().Get("Vary")
			if gotVary != tc.expectedVary {
				t.Errorf("expected Vary %q, got %q", tc.expectedVary, gotVary)
			}

			// For matching GET requests, ensure dummy handler was called
			if tc.expectedStatus == http.StatusOK && tc.method != http.MethodOptions {
				if rr.Body.String() != "OK" {
					t.Errorf("expected body 'OK', got %q", rr.Body.String())
				}
			}
		})
	}
}
