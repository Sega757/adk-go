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
		origin          string
		method          string
		wantStatus      int
		wantAllowOrigin string
		wantVary        string
	}{
		{
			name:            "no origin header",
			frontendAddress: "localhost:8080",
			origin:          "",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "",
			wantVary:        "",
		},
		{
			name:            "matching origin - scheme-less HTTP",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "http://localhost:8080",
			wantVary:        "Origin",
		},
		{
			name:            "matching origin - scheme-less HTTPS",
			frontendAddress: "localhost:8080",
			origin:          "https://localhost:8080",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "https://localhost:8080",
			wantVary:        "Origin",
		},
		{
			name:            "matching origin - exact schemed HTTP",
			frontendAddress: "http://example.com",
			origin:          "http://example.com",
			method:          "GET",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "http://example.com",
			wantVary:        "Origin",
		},
		{
			name:            "mismatching origin - exact schemed HTTP tries HTTPS",
			frontendAddress: "http://example.com",
			origin:          "https://example.com",
			method:          "GET",
			wantStatus:      http.StatusForbidden,
		},
		{
			name:            "mismatching origin",
			frontendAddress: "localhost:8080",
			origin:          "http://malicious.com",
			method:          "GET",
			wantStatus:      http.StatusForbidden,
		},
		{
			name:            "preflight OPTIONS with matching origin",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "OPTIONS",
			wantStatus:      http.StatusOK,
			wantAllowOrigin: "http://localhost:8080",
			wantVary:        "Origin",
		},
		{
			name:            "preflight OPTIONS with mismatching origin",
			frontendAddress: "localhost:8080",
			origin:          "http://malicious.com",
			method:          "OPTIONS",
			wantStatus:      http.StatusForbidden,
		},
		{
			name:            "empty frontend address rejects all origins",
			frontendAddress: "",
			origin:          "http://localhost:8080",
			method:          "GET",
			wantStatus:      http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := corsWithArgs(tt.frontendAddress)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/any-path", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantStatus == http.StatusOK {
				gotAllowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
				if gotAllowOrigin != tt.wantAllowOrigin {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.wantAllowOrigin, gotAllowOrigin)
				}
				gotVary := rr.Header().Get("Vary")
				if gotVary != tt.wantVary {
					t.Errorf("expected Vary %q, got %q", tt.wantVary, gotVary)
				}
			}
		})
	}
}
