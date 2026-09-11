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

package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewErrorHandler(t *testing.T) {
	tests := []struct {
		name       string
		handlerErr error
		wantCode   int
		wantBody   string
	}{
		{
			name:       "status error non-500 preserved",
			handlerErr: newStatusError(errors.New("bad request param"), http.StatusBadRequest),
			wantCode:   http.StatusBadRequest,
			wantBody:   "bad request param\n",
		},
		{
			name:       "status error 500 sanitized",
			handlerErr: newStatusError(errors.New("secret db connection failed"), http.StatusInternalServerError),
			wantCode:   http.StatusInternalServerError,
			wantBody:   "internal server error\n",
		},
		{
			name:       "standard error 500 sanitized",
			handlerErr: errors.New("sensitive stack trace or internal error"),
			wantCode:   http.StatusInternalServerError,
			wantBody:   "internal server error\n",
		},
		{
			name:       "no error",
			handlerErr: nil,
			wantCode:   http.StatusOK,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
				return tt.handlerErr
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantCode)
			}
			if gotBody := rec.Body.String(); gotBody != tt.wantBody {
				t.Errorf("got body %q, want %q", gotBody, tt.wantBody)
			}
		})
	}
}

func TestEncodeJSONResponse(t *testing.T) {
	t.Run("successful encoding", func(t *testing.T) {
		rec := httptest.NewRecorder()
		data := map[string]string{"status": "ok"}
		EncodeJSONResponse(data, http.StatusOK, rec)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		if gotBody := strings.TrimSpace(rec.Body.String()); gotBody != `{"status":"ok"}` {
			t.Errorf("got body %q, want %q", gotBody, `{"status":"ok"}`)
		}
	})

	t.Run("encoding failure sanitization", func(t *testing.T) {
		// Channels cannot be marshaled to JSON and will trigger an encoding error.
		unsupportedType := make(chan int)

		rec := httptest.NewRecorder()
		EncodeJSONResponse(unsupportedType, http.StatusOK, rec)

		// When JSON encoding fails after WriteHeader, http.Error writes sanitized error text
		if gotBody := rec.Body.String(); !strings.Contains(gotBody, "internal server error") {
			t.Errorf("got body %q, expected to contain 'internal server error'", gotBody)
		}
	})
}
