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
		requestMethod   string
		requestOrigin   string
		expectedStatus  int
		expectCORS      bool
	}{
		{
			name:            "No Origin header - request should proceed",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "",
			expectedStatus:  http.StatusOK,
			expectCORS:      false,
		},
		{
			name:            "Correct origin, scheme-less config, http",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "http://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectCORS:      true,
		},
		{
			name:            "Correct origin, scheme-less config, https",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "https://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectCORS:      true,
		},
		{
			name:            "Incorrect origin host",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "http://malicious.com",
			expectedStatus:  http.StatusForbidden,
			expectCORS:      false,
		},
		{
			name:            "Invalid origin header",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "http://[::1]:foo", // invalid host for url.Parse
			expectedStatus:  http.StatusForbidden,
			expectCORS:      false,
		},
		{
			name:            "Schemed config - correct origin",
			frontendAddress: "https://localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "https://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectCORS:      true,
		},
		{
			name:            "Schemed config - scheme mismatch",
			frontendAddress: "https://localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "http://localhost:8080",
			expectedStatus:  http.StatusForbidden,
			expectCORS:      false,
		},
		{
			name:            "Preflight OPTIONS request with correct origin",
			frontendAddress: "localhost:8080",
			requestMethod:   "OPTIONS",
			requestOrigin:   "http://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectCORS:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			innerCalled := false
			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				innerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := corsWithArgs(tc.frontendAddress)
			handler := middleware(innerHandler)

			req := httptest.NewRequest(tc.requestMethod, "/api/some-endpoint", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			if tc.expectCORS {
				if w.Header().Get("Access-Control-Allow-Origin") != tc.requestOrigin {
					t.Errorf("expected Access-Control-Allow-Origin to be %q, got %q", tc.requestOrigin, w.Header().Get("Access-Control-Allow-Origin"))
				}
				if w.Header().Get("Vary") != "Origin" {
					t.Errorf("expected Vary to be 'Origin', got %q", w.Header().Get("Vary"))
				}
				if tc.requestMethod == "OPTIONS" && innerCalled {
					t.Error("expected inner handler not to be called for OPTIONS preflight request")
				}
			} else {
				if w.Header().Get("Access-Control-Allow-Origin") != "" {
					t.Errorf("unexpected Access-Control-Allow-Origin header: %q", w.Header().Get("Access-Control-Allow-Origin"))
				}
			}
		})
	}
}
