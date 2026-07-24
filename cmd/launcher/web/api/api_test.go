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
		requestOrigin   string
		expectedStatus  int
		expectedAllow   string
		expectedVary    bool
	}{
		{
			name:            "scheme-less: allowed origin HTTP",
			frontendAddress: "localhost:8080",
			method:          "GET",
			requestOrigin:   "http://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "http://localhost:8080",
			expectedVary:    true,
		},
		{
			name:            "scheme-less: allowed origin HTTPS",
			frontendAddress: "localhost:8080",
			method:          "GET",
			requestOrigin:   "https://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "https://localhost:8080",
			expectedVary:    true,
		},
		{
			name:            "scheme-less: disallowed origin",
			frontendAddress: "localhost:8080",
			method:          "GET",
			requestOrigin:   "http://unauthorized.com",
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    true,
		},
		{
			name:            "scheme-less: invalid origin",
			frontendAddress: "localhost:8080",
			method:          "GET",
			requestOrigin:   "://invalid-url",
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    true,
		},
		{
			name:            "scheme-less: missing origin (direct client)",
			frontendAddress: "localhost:8080",
			method:          "GET",
			requestOrigin:   "",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "",
			expectedVary:    false,
		},
		{
			name:            "schemed: allowed exact match",
			frontendAddress: "https://safe.example.com",
			method:          "GET",
			requestOrigin:   "https://safe.example.com",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "https://safe.example.com",
			expectedVary:    true,
		},
		{
			name:            "schemed: disallowed scheme (CORS bypass prevention)",
			frontendAddress: "https://safe.example.com",
			method:          "GET",
			requestOrigin:   "http://safe.example.com",
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    true,
		},
		{
			name:            "schemed: disallowed host",
			frontendAddress: "https://safe.example.com",
			method:          "GET",
			requestOrigin:   "https://unsafe.example.com",
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    true,
		},
		{
			name:            "preflight OPTIONS matched",
			frontendAddress: "localhost:8080",
			method:          "OPTIONS",
			requestOrigin:   "http://localhost:8080",
			expectedStatus:  http.StatusOK,
			expectedAllow:   "http://localhost:8080",
			expectedVary:    true,
		},
		{
			name:            "preflight OPTIONS unmatched",
			frontendAddress: "localhost:8080",
			method:          "OPTIONS",
			requestOrigin:   "http://unauthorized.com",
			expectedStatus:  http.StatusForbidden,
			expectedAllow:   "",
			expectedVary:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create dummy handler
			handlerCalled := false
			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with CORS middleware
			corsMiddleware := corsWithArgs(tc.frontendAddress)
			router := corsMiddleware(dummyHandler)

			// Execute request
			req := httptest.NewRequest(tc.method, "/api/test", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Assertions
			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			allowHeader := rr.Header().Get("Access-Control-Allow-Origin")
			if allowHeader != tc.expectedAllow {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tc.expectedAllow, allowHeader)
			}

			varyHeader := rr.Header().Get("Vary")
			hasVaryOrigin := varyHeader == "Origin"
			if hasVaryOrigin != tc.expectedVary {
				t.Errorf("expected Vary header Presence %t, got header value %q", tc.expectedVary, varyHeader)
			}

			// If status is OK and method is not OPTIONS, dummy handler should have been called
			expectedHandlerCalled := tc.expectedStatus == http.StatusOK && tc.method != "OPTIONS"
			if handlerCalled != expectedHandlerCalled {
				t.Errorf("expected handler called to be %t, got %t", expectedHandlerCalled, handlerCalled)
			}
		})
	}
}
