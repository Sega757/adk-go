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
		name                 string
		frontendAddress      string
		requestOrigin        string
		requestMethod        string
		wantAllowedOrigin    string
		wantMethodsHeader    bool
		wantVaryHeader       bool
		wantStatus           int
	}{
		{
			name:              "Allowed origin with schemed frontend address",
			frontendAddress:   "http://localhost:8080",
			requestOrigin:     "http://localhost:8080",
			requestMethod:     "GET",
			wantAllowedOrigin: "http://localhost:8080",
			wantMethodsHeader: true,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Disallowed origin with schemed frontend address",
			frontendAddress:   "http://localhost:8080",
			requestOrigin:     "http://attacker.com",
			requestMethod:     "GET",
			wantAllowedOrigin: "",
			wantMethodsHeader: false,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Scheme-less frontend address matching http",
			frontendAddress:   "localhost:8080",
			requestOrigin:     "http://localhost:8080",
			requestMethod:     "GET",
			wantAllowedOrigin: "http://localhost:8080",
			wantMethodsHeader: true,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Scheme-less frontend address matching https",
			frontendAddress:   "localhost:8080",
			requestOrigin:     "https://localhost:8080",
			requestMethod:     "GET",
			wantAllowedOrigin: "https://localhost:8080",
			wantMethodsHeader: true,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Scheme-less frontend address mismatch port",
			frontendAddress:   "localhost:8080",
			requestOrigin:     "http://localhost:9090",
			requestMethod:     "GET",
			wantAllowedOrigin: "",
			wantMethodsHeader: false,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Preflight OPTIONS request with valid origin",
			frontendAddress:   "http://localhost:8080",
			requestOrigin:     "http://localhost:8080",
			requestMethod:     "OPTIONS",
			wantAllowedOrigin: "http://localhost:8080",
			wantMethodsHeader: true,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Preflight OPTIONS request with invalid origin",
			frontendAddress:   "http://localhost:8080",
			requestOrigin:     "http://malicious.com",
			requestMethod:     "OPTIONS",
			wantAllowedOrigin: "",
			wantMethodsHeader: false,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Missing origin header",
			frontendAddress:   "http://localhost:8080",
			requestOrigin:     "",
			requestMethod:     "GET",
			wantAllowedOrigin: "",
			wantMethodsHeader: false,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Disallowed scheme for scheme-less frontend address",
			frontendAddress:   "localhost:8080",
			requestOrigin:     "ftp://localhost:8080",
			requestMethod:     "GET",
			wantAllowedOrigin: "",
			wantMethodsHeader: false,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
		{
			name:              "Scheme mismatch when frontend address is schemed",
			frontendAddress:   "https://localhost:8080",
			requestOrigin:     "http://localhost:8080",
			requestMethod:     "GET",
			wantAllowedOrigin: "",
			wantMethodsHeader: false,
			wantVaryHeader:    true,
			wantStatus:        http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := corsWithArgs(tt.frontendAddress)(dummyHandler)

			req := httptest.NewRequest(tt.requestMethod, "/api/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantVaryHeader {
				if got := rec.Header().Get("Vary"); got != "Origin" {
					t.Errorf("got Vary header %q, want %q", got, "Origin")
				}
			}

			gotOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if gotOrigin != tt.wantAllowedOrigin {
				t.Errorf("got Access-Control-Allow-Origin %q, want %q", gotOrigin, tt.wantAllowedOrigin)
			}

			gotMethods := rec.Header().Get("Access-Control-Allow-Methods")
			if tt.wantMethodsHeader && gotMethods == "" {
				t.Errorf("expected Access-Control-Allow-Methods header, got empty")
			} else if !tt.wantMethodsHeader && gotMethods != "" {
				t.Errorf("unexpected Access-Control-Allow-Methods header: %q", gotMethods)
			}
		})
	}
}
