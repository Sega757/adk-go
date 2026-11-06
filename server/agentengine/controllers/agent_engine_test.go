// Copyright 2026 Google LLC
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

package controllers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"google.golang.org/adk/v2/server/agentengine/controllers"
	"google.golang.org/adk/v2/server/agentengine/controllers/method"
	"google.golang.org/adk/v2/session"
)

type dummyHandler struct {
	calledWith []byte
}

func (d *dummyHandler) Name() string { return "test_method" }

func (d *dummyHandler) Handle(_ context.Context, rw http.ResponseWriter, payload []byte) error {
	d.calledWith = append([]byte(nil), payload...)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(`{"status":"ok"}`))
	return nil
}

func (d *dummyHandler) Metadata() (*structpb.Struct, error) {
	return nil, nil
}

func TestQueryPayloadLimit(t *testing.T) {
	t.Parallel()

	const limit = 100

	dummy := &dummyHandler{}
	ctrl, err := controllers.NewAgentEngineAPIController(
		session.InMemoryService(),
		time.Second,
		limit,
		[]method.MethodHandler{dummy},
	)
	if err != nil {
		t.Fatalf("NewAgentEngineAPIController failed: %v", err)
	}

	t.Run("within payload limit succeeds", func(t *testing.T) {
		body := `{"class_method":"test_method"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rec := httptest.NewRecorder()

		ctrl.Query(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if string(dummy.calledWith) != body {
			t.Fatalf("handler payload = %q, want %q", string(dummy.calledWith), body)
		}
	})

	t.Run("exceeding payload limit returns 413 Request Entity Too Large", func(t *testing.T) {
		body := `{"class_method":"test_method","padding":"` + strings.Repeat("x", limit+10) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rec := httptest.NewRecorder()

		ctrl.Query(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "http: request body too large") {
			t.Fatalf("body = %q, want mention of 'http: request body too large'", rec.Body.String())
		}
	})
}
