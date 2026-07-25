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
		method          string
		origin          string
		wantStatus      int
		wantCORSHeaders bool
	}{
		{
			name:            "No Origin Header Bypass",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "",
			wantStatus:      http.StatusOK,
			wantCORSHeaders: false,
		},
		{
			name:            "Schemed Allowed Match",
			frontendAddress: "http://localhost:8080",
			method:          http.MethodGet,
			origin:          "http://localhost:8080",
			wantStatus:      http.StatusOK,
			wantCORSHeaders: true,
		},
		{
			name:            "Schemed Mismatched Scheme Rejected",
			frontendAddress: "http://localhost:8080",
			method:          http.MethodGet,
			origin:          "https://localhost:8080",
			wantStatus:      http.StatusForbidden,
			wantCORSHeaders: false,
		},
		{
			name:            "Schemed Mismatched Host Rejected",
			frontendAddress: "http://localhost:8080",
			method:          http.MethodGet,
			origin:          "http://evil.com",
			wantStatus:      http.StatusForbidden,
			wantCORSHeaders: false,
		},
		{
			name:            "SchemeLess Allowed HTTP",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "http://localhost:8080",
			wantStatus:      http.StatusOK,
			wantCORSHeaders: true,
		},
		{
			name:            "SchemeLess Allowed HTTPS",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "https://localhost:8080",
			wantStatus:      http.StatusOK,
			wantCORSHeaders: true,
		},
		{
			name:            "SchemeLess Mismatched Host Rejected",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "http://evil.com",
			wantStatus:      http.StatusForbidden,
			wantCORSHeaders: false,
		},
		{
			name:            "Invalid Origin Format Rejected",
			frontendAddress: "localhost:8080",
			method:          http.MethodGet,
			origin:          "://invalid-url",
			wantStatus:      http.StatusForbidden,
			wantCORSHeaders: false,
		},
		{
			name:            "OPTIONS Preflight Allowed Match",
			frontendAddress: "http://localhost:8080",
			method:          http.MethodOptions,
			origin:          "http://localhost:8080",
			wantStatus:      http.StatusOK,
			wantCORSHeaders: true,
		},
		{
			name:            "OPTIONS Preflight Mismatched Rejected",
			frontendAddress: "http://localhost:8080",
			method:          http.MethodOptions,
			origin:          "http://evil.com",
			wantStatus:      http.StatusForbidden,
			wantCORSHeaders: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a dummy inner handler that returns 200 OK and "InnerOK"
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("InnerOK"))
			})

			middleware := corsWithArgs(tc.frontendAddress)(inner)

			req := httptest.NewRequest(tc.method, "/test", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}

			if tc.wantCORSHeaders {
				if got := rr.Header().Get("Access-Control-Allow-Origin"); got != tc.origin {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tc.origin, got)
				}
				if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
					t.Error("expected Access-Control-Allow-Methods to be set")
				}
				if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
					t.Error("expected Access-Control-Allow-Headers to be set")
				}
				if got := rr.Header().Get("Vary"); got != "Origin" {
					t.Errorf("expected Vary header %q, got %q", "Origin", got)
				}
			} else {
				if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Errorf("expected no Access-Control-Allow-Origin, got %q", got)
				}
			}
		})
	}
}
