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

func TestCORSWithArgs(t *testing.T) {
	tests := []struct {
		name            string
		frontendAddress string
		requestMethod   string
		requestOrigin   string
		expectCORS      bool
		expectOrigin    string
		expectStatus    int
		expectServed    bool
	}{
		{
			name:            "scheme-less matches http",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "http://localhost:8080",
			expectCORS:      true,
			expectOrigin:    "http://localhost:8080",
			expectStatus:    http.StatusOK,
			expectServed:    true,
		},
		{
			name:            "scheme-less matches https",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "https://localhost:8080",
			expectCORS:      true,
			expectOrigin:    "https://localhost:8080",
			expectStatus:    http.StatusOK,
			expectServed:    true,
		},
		{
			name:            "scheme-less trailing slash matches",
			frontendAddress: "localhost:8080/",
			requestMethod:   "GET",
			requestOrigin:   "https://localhost:8080/",
			expectCORS:      true,
			expectOrigin:    "https://localhost:8080/",
			expectStatus:    http.StatusOK,
			expectServed:    true,
		},
		{
			name:            "scheme-less mismatch domain",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "http://evil-localhost:8080",
			expectCORS:      false,
			expectStatus:    http.StatusOK,
			expectServed:    true,
		},
		{
			name:            "schemed exact match",
			frontendAddress: "https://myhost.com",
			requestMethod:   "GET",
			requestOrigin:   "https://myhost.com",
			expectCORS:      true,
			expectOrigin:    "https://myhost.com",
			expectStatus:    http.StatusOK,
			expectServed:    true,
		},
		{
			name:            "schemed mismatch scheme",
			frontendAddress: "https://myhost.com",
			requestMethod:   "GET",
			requestOrigin:   "http://myhost.com",
			expectCORS:      false,
			expectStatus:    http.StatusOK,
			expectServed:    true,
		},
		{
			name:            "empty origin omitted",
			frontendAddress: "localhost:8080",
			requestMethod:   "GET",
			requestOrigin:   "",
			expectCORS:      false,
			expectStatus:    http.StatusOK,
			expectServed:    true,
		},
		{
			name:            "OPTIONS preflight success",
			frontendAddress: "localhost:8080",
			requestMethod:   "OPTIONS",
			requestOrigin:   "http://localhost:8080",
			expectCORS:      true,
			expectOrigin:    "http://localhost:8080",
			expectStatus:    http.StatusOK,
			expectServed:    false, // OPTIONS is intercepted and handled by CORS
		},
		{
			name:            "OPTIONS preflight mismatch",
			frontendAddress: "localhost:8080",
			requestMethod:   "OPTIONS",
			requestOrigin:   "http://evil-localhost:8080",
			expectCORS:      false,
			expectStatus:    http.StatusOK,
			expectServed:    true, // proceeds to next handler
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			served := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				served = true
				w.WriteHeader(http.StatusOK)
			})

			handler := corsWithArgs(tc.frontendAddress)(next)

			req := httptest.NewRequest(tc.requestMethod, "/api/test", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check Vary header is set correctly
			vary := rr.Header().Get("Vary")
			if vary != "Origin" {
				t.Errorf("expected Vary header 'Origin', got '%s'", vary)
			}

			if tc.expectCORS {
				gotOrigin := rr.Header().Get("Access-Control-Allow-Origin")
				if gotOrigin != tc.expectOrigin {
					t.Errorf("expected Access-Control-Allow-Origin '%s', got '%s'", tc.expectOrigin, gotOrigin)
				}
				methods := rr.Header().Get("Access-Control-Allow-Methods")
				if methods == "" {
					t.Errorf("expected Access-Control-Allow-Methods to be set")
				}
				headers := rr.Header().Get("Access-Control-Allow-Headers")
				if headers == "" {
					t.Errorf("expected Access-Control-Allow-Headers to be set")
				}
			} else {
				gotOrigin := rr.Header().Get("Access-Control-Allow-Origin")
				if gotOrigin != "" {
					t.Errorf("expected no Access-Control-Allow-Origin header, got '%s'", gotOrigin)
				}
			}

			if rr.Code != tc.expectStatus {
				t.Errorf("expected status %d, got %d", tc.expectStatus, rr.Code)
			}

			if served != tc.expectServed {
				t.Errorf("expected served to be %v, got %v", tc.expectServed, served)
			}
		})
	}
}
