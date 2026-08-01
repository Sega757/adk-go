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

func TestCorsWithArgs(t *testing.T) {
	tests := []struct {
		name            string
		frontendAddress string
		origin          string
		method          string
		expectStatus    int
		expectCORS      bool
		expectNextRun   bool
	}{
		{
			name:            "No Origin header - pass through",
			frontendAddress: "localhost:8080",
			origin:          "",
			method:          "GET",
			expectStatus:    http.StatusOK,
			expectCORS:      false,
			expectNextRun:   true,
		},
		{
			name:            "Invalid Origin header - missing scheme",
			frontendAddress: "localhost:8080",
			origin:          "localhost:8080", // missing scheme
			method:          "GET",
			expectStatus:    http.StatusForbidden,
			expectCORS:      false,
			expectNextRun:   false,
		},
		{
			name:            "Invalid Origin header - empty host",
			frontendAddress: "localhost:8080",
			origin:          "http://", // empty host
			method:          "GET",
			expectStatus:    http.StatusForbidden,
			expectCORS:      false,
			expectNextRun:   false,
		},
		{
			name:            "Scheme-less config - matching origin with http",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "GET",
			expectStatus:    http.StatusOK,
			expectCORS:      true,
			expectNextRun:   true,
		},
		{
			name:            "Scheme-less config - matching origin with https",
			frontendAddress: "localhost:8080",
			origin:          "https://localhost:8080",
			method:          "GET",
			expectStatus:    http.StatusOK,
			expectCORS:      true,
			expectNextRun:   true,
		},
		{
			name:            "Scheme-less config - mismatched origin",
			frontendAddress: "localhost:8080",
			origin:          "http://attacker.com:8080",
			method:          "GET",
			expectStatus:    http.StatusForbidden,
			expectCORS:      false,
			expectNextRun:   false,
		},
		{
			name:            "Schemed config - matching origin and scheme",
			frontendAddress: "https://example.com",
			origin:          "https://example.com",
			method:          "GET",
			expectStatus:    http.StatusOK,
			expectCORS:      true,
			expectNextRun:   true,
		},
		{
			name:            "Schemed config - mismatched scheme",
			frontendAddress: "https://example.com",
			origin:          "http://example.com",
			method:          "GET",
			expectStatus:    http.StatusForbidden,
			expectCORS:      false,
			expectNextRun:   false,
		},
		{
			name:            "Schemed config - mismatched host",
			frontendAddress: "https://example.com",
			origin:          "https://attacker.com",
			method:          "GET",
			expectStatus:    http.StatusForbidden,
			expectCORS:      false,
			expectNextRun:   false,
		},
		{
			name:            "Preflight OPTIONS request - valid origin",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "OPTIONS",
			expectStatus:    http.StatusOK,
			expectCORS:      true,
			expectNextRun:   false, // OPTIONS should short-circuit and not call next
		},
		{
			name:            "Preflight OPTIONS request - invalid origin",
			frontendAddress: "localhost:8080",
			origin:          "http://attacker.com",
			method:          "OPTIONS",
			expectStatus:    http.StatusForbidden,
			expectCORS:      false,
			expectNextRun:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextRun := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextRun = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := corsWithArgs(tc.frontendAddress)
			handler := middleware(nextHandler)

			req := httptest.NewRequest(tc.method, "/api/list-apps", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tc.expectStatus {
				t.Errorf("expected status %d, got %d", tc.expectStatus, w.Code)
			}

			if tc.expectCORS {
				gotOrigin := w.Header().Get("Access-Control-Allow-Origin")
				if gotOrigin != tc.origin {
					t.Errorf("expected Access-Control-Allow-Origin to be %q, got %q", tc.origin, gotOrigin)
				}
				gotMethods := w.Header().Get("Access-Control-Allow-Methods")
				if gotMethods == "" {
					t.Errorf("expected Access-Control-Allow-Methods to be set")
				}
				gotHeaders := w.Header().Get("Access-Control-Allow-Headers")
				if gotHeaders == "" {
					t.Errorf("expected Access-Control-Allow-Headers to be set")
				}
				gotVary := w.Header().Get("Vary")
				if gotVary != "Origin" {
					t.Errorf("expected Vary header to be 'Origin', got %q", gotVary)
				}
			} else {
				gotOrigin := w.Header().Get("Access-Control-Allow-Origin")
				if gotOrigin != "" {
					t.Errorf("expected no Access-Control-Allow-Origin, got %q", gotOrigin)
				}
			}

			if nextRun != tc.expectNextRun {
				t.Errorf("expected next handler execution to be %t, got %t", tc.expectNextRun, nextRun)
			}
		})
	}
}
