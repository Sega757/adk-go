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
		method          string
		origin          string
		expectedStatus  int
		expectedOrigin  string
		expectedVary    string
	}{
		{
			name:            "Missing Origin header is allowed (not a cross-origin request)",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "",
			expectedStatus:  http.StatusOK,
			expectedOrigin:  "",
			expectedVary:    "",
		},
		{
			name:            "Scheme-less frontend address, matching HTTP origin",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "http://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedOrigin:  "http://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "Scheme-less frontend address, matching HTTPS origin",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "https://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedOrigin:  "https://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "Scheme-less frontend address, non-matching host rejected",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "http://attacker.com",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "Scheme-less frontend address, non-http/https scheme rejected",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "ftp://localhost:8080",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "Schemed frontend address (HTTP), matching HTTP origin",
			frontendAddress: "http://localhost:8080",
			method:          http.MethodGet,
			origin:          "http://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedOrigin:  "http://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "Schemed frontend address (HTTP), mismatching HTTPS origin rejected",
			frontendAddress: "http://localhost:8080",
			method:          http.MethodGet,
			origin:          "https://localhost:8080",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "Schemed frontend address (HTTPS), matching HTTPS origin",
			frontendAddress: "https://localhost:8080",
			method:          http.MethodGet,
			origin:          "https://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedOrigin:  "https://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "Schemed frontend address (HTTPS), mismatching HTTP origin rejected",
			frontendAddress: "https://localhost:8080",
			method:          http.MethodGet,
			origin:          "http://localhost:8080",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "Invalid origin header rejected",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "://invalid-origin",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "OPTIONS Preflight request with valid matching origin",
			frontendAddress: "localhost:8080",
			method:          http.MethodOptions,
			origin:          "http://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedOrigin:  "http://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "OPTIONS Preflight request with invalid origin",
			frontendAddress: "localhost:8080",
			method:          http.MethodOptions,
			origin:          "http://attacker.com",
			expectedStatus:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			})

			middleware := corsWithArgs(tt.frontendAddress)
			handler := middleware(innerHandler)

			req := httptest.NewRequest(tt.method, "http://localhost:8080/api/list-apps", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				gotOrigin := rec.Header().Get("Access-Control-Allow-Origin")
				if gotOrigin != tt.expectedOrigin {
					t.Errorf("expected Access-Control-Allow-Origin to be %q, got %q", tt.expectedOrigin, gotOrigin)
				}

				gotVary := rec.Header().Get("Vary")
				if gotVary != tt.expectedVary {
					t.Errorf("expected Vary to be %q, got %q", tt.expectedVary, gotVary)
				}
			}
		})
	}
}
