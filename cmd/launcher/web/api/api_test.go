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
		wantOrigin      string
		wantVary        bool
		wantStatus      int
	}{
		{
			name:            "Localhost scheme-less config with http origin",
			frontendAddress: "localhost:8080",
			requestMethod:   http.MethodGet,
			requestOrigin:   "http://localhost:8080",
			wantOrigin:      "http://localhost:8080",
			wantVary:        true,
			wantStatus:      http.StatusOK,
		},
		{
			name:            "Localhost scheme-less config with https origin",
			frontendAddress: "localhost:8080",
			requestMethod:   http.MethodGet,
			requestOrigin:   "https://localhost:8080",
			wantOrigin:      "https://localhost:8080",
			wantVary:        true,
			wantStatus:      http.StatusOK,
		},
		{
			name:            "Localhost scheme-less config with malicious origin",
			frontendAddress: "localhost:8080",
			requestMethod:   http.MethodGet,
			requestOrigin:   "http://malicious.com",
			wantOrigin:      "",
			wantVary:        false,
			wantStatus:      http.StatusOK,
		},
		{
			name:            "Full scheme config matching",
			frontendAddress: "https://example.com",
			requestMethod:   http.MethodGet,
			requestOrigin:   "https://example.com",
			wantOrigin:      "https://example.com",
			wantVary:        true,
			wantStatus:      http.StatusOK,
		},
		{
			name:            "Full scheme config with mismatched scheme request",
			frontendAddress: "https://example.com",
			requestMethod:   http.MethodGet,
			requestOrigin:   "http://example.com",
			wantOrigin:      "",
			wantVary:        false,
			wantStatus:      http.StatusOK,
		},
		{
			name:            "No Origin header fallback",
			frontendAddress: "localhost:8080",
			requestMethod:   http.MethodGet,
			requestOrigin:   "",
			wantOrigin:      "http://localhost:8080",
			wantVary:        false,
			wantStatus:      http.StatusOK,
		},
		{
			name:            "OPTIONS request handling",
			frontendAddress: "localhost:8080",
			requestMethod:   http.MethodOptions,
			requestOrigin:   "http://localhost:8080",
			wantOrigin:      "http://localhost:8080",
			wantVary:        true,
			wantStatus:      http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := corsWithArgs(tt.frontendAddress)
			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			r := httptest.NewRequest(tt.requestMethod, "/api/test", nil)
			if tt.requestOrigin != "" {
				r.Header.Set("Origin", tt.requestOrigin)
			}
			w := httptest.NewRecorder()

			middleware(nextHandler).ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			gotOrigin := res.Header.Get("Access-Control-Allow-Origin")
			if gotOrigin != tt.wantOrigin {
				t.Errorf("got Access-Control-Allow-Origin = %q, want %q", gotOrigin, tt.wantOrigin)
			}

			gotVary := res.Header.Get("Vary") == "Origin"
			if gotVary != tt.wantVary {
				t.Errorf("got Vary: Origin header presence = %t, want %t", gotVary, tt.wantVary)
			}

			if res.StatusCode != tt.wantStatus {
				t.Errorf("got Status = %d, want %d", res.StatusCode, tt.wantStatus)
			}

			if tt.requestMethod == http.MethodOptions {
				if handlerCalled {
					t.Error("expected next handler not to be called for OPTIONS request")
				}
			} else {
				if !handlerCalled {
					t.Error("expected next handler to be called")
				}
			}
		})
	}
}
