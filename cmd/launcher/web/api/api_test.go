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
		expectedStatus  int
		expectedAllow   string
		expectedVary    string
	}{
		{
			name:            "scheme-less allowed origin matching http",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "http://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "scheme-less allowed origin matching https",
			frontendAddress: "localhost:8080",
			origin:          "https://localhost:8080",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "https://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "scheme-less allowed origin mismatch",
			frontendAddress: "localhost:8080",
			origin:          "http://otherhost:8080",
			method:          "GET",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "schemed allowed origin matching http",
			frontendAddress: "http://localhost:8080",
			origin:          "http://localhost:8080",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "http://localhost:8080",
			expectedVary:    "Origin",
		},
		{
			name:            "schemed allowed origin cross-scheme mismatch (http allowed, https requested)",
			frontendAddress: "http://localhost:8080",
			origin:          "https://localhost:8080",
			method:          "GET",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "schemed allowed origin cross-scheme mismatch (https allowed, http requested)",
			frontendAddress: "https://localhost:8080",
			origin:          "http://localhost:8080",
			method:          "GET",
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "missing origin header allowed to pass through",
			frontendAddress: "localhost:8080",
			origin:          "",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "",
		},
		{
			name:            "preflight request success",
			frontendAddress: "localhost:8080",
			origin:          "http://localhost:8080",
			method:          "OPTIONS",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "http://localhost:8080",
			expectedVary:    "Origin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := corsWithArgs(tc.frontendAddress)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			}))

			req := httptest.NewRequest(tc.method, "/any-path", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				gotAllow := rr.Header().Get("Access-Control-Allow-Origin")
				if gotAllow != tc.expectedAllow {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tc.expectedAllow, gotAllow)
				}

				if tc.origin != "" {
					gotVary := rr.Header().Get("Vary")
					if gotVary != tc.expectedVary {
						t.Errorf("expected Vary header %q, got %q", tc.expectedVary, gotVary)
					}
				}
			}
		})
	}
}
