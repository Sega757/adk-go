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
		wantStatus      int
		wantAllowOrigin string
		wantVary        string
	}{
		{
			name:            "scheme-less frontend address, http match",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "http://localhost:8080",
			wantVary:        "Origin",
		},
		{
			name:            "scheme-less frontend address, https match",
			frontendAddress: "localhost:8080",
			origin:          "https://localhost:8080",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "https://localhost:8080",
			wantVary:        "Origin",
		},
		{
			name:            "scheme-less frontend address, mismatch rejection",
			frontendAddress: "localhost:8080",
			origin:          "http://evil.com",
			method:          "GET",
			wantStatus:      http.StatusForbidden,
		},
		{
			name:            "schemed frontend address, exact http match",
			frontendAddress: "http://localhost:8080",
			origin:          "http://localhost:8080",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "http://localhost:8080",
			wantVary:        "Origin",
		},
		{
			name:            "schemed frontend address, https mismatch rejection",
			frontendAddress: "http://localhost:8080",
			origin:          "https://localhost:8080",
			method:          "GET",
			wantStatus:      http.StatusForbidden,
		},
		{
			name:            "no Origin header, falls back to setting configured frontendAddress",
			frontendAddress: "localhost:8080",
			origin:          "",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "localhost:8080",
			wantVary:        "",
		},
		{
			name:            "OPTIONS preflight request, returns OK",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "OPTIONS",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "http://localhost:8080",
			wantVary:        "Origin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := corsWithArgs(tc.frontendAddress)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tc.method, "/some-endpoint", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status code %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantStatus == http.StatusOK {
				allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
				if allowOrigin != tc.wantAllowOrigin {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tc.wantAllowOrigin, allowOrigin)
				}
				vary := rec.Header().Get("Vary")
				if vary != tc.wantVary {
					t.Errorf("expected Vary %q, got %q", tc.wantVary, vary)
				}
				allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
				if allowMethods != "GET, POST, PUT, DELETE, OPTIONS" {
					t.Errorf("expected Access-Control-Allow-Methods, got %q", allowMethods)
				}
			}
		})
	}
}
