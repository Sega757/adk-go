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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"google.golang.org/adk/v2/artifact"
	"google.golang.org/adk/v2/server/adkrest/controllers"
)

type fakeArtifactService struct {
	artifact.Service
	err error
}

func (f *fakeArtifactService) List(ctx context.Context, req *artifact.ListRequest) (*artifact.ListResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &artifact.ListResponse{FileNames: []string{"test.txt"}}, nil
}

func (f *fakeArtifactService) Load(ctx context.Context, req *artifact.LoadRequest) (*artifact.LoadResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &artifact.LoadResponse{}, nil
}

func (f *fakeArtifactService) Delete(ctx context.Context, req *artifact.DeleteRequest) error {
	return f.err
}

func TestArtifactsAPIController_InternalServerErrorSanitized(t *testing.T) {
	svc := &fakeArtifactService{err: errors.New("secret db connection string failed")}
	controller := controllers.NewArtifactsAPIController(svc)

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		vars    map[string]string
		path    string
	}{
		{
			name:    "ListArtifactsHandler",
			handler: controller.ListArtifactsHandler,
			vars:    map[string]string{"app_name": "app", "user_id": "user", "session_id": "session"},
			path:    "/apps/app/users/user/sessions/session/artifacts",
		},
		{
			name:    "LoadArtifactHandler",
			handler: controller.LoadArtifactHandler,
			vars:    map[string]string{"app_name": "app", "user_id": "user", "session_id": "session", "artifact_name": "test.txt"},
			path:    "/apps/app/users/user/sessions/session/artifacts/test.txt",
		},
		{
			name:    "LoadArtifactVersionHandler",
			handler: controller.LoadArtifactVersionHandler,
			vars:    map[string]string{"app_name": "app", "user_id": "user", "session_id": "session", "artifact_name": "test.txt", "version": "1"},
			path:    "/apps/app/users/user/sessions/session/artifacts/test.txt/versions/1",
		},
		{
			name:    "DeleteArtifactHandler",
			handler: controller.DeleteArtifactHandler,
			vars:    map[string]string{"app_name": "app", "user_id": "user", "session_id": "session", "artifact_name": "test.txt"},
			path:    "/apps/app/users/user/sessions/session/artifacts/test.txt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req = mux.SetURLVars(req, tc.vars)
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("got status code %d, want %d", rec.Code, http.StatusInternalServerError)
			}

			body := strings.TrimSpace(rec.Body.String())
			if body != "internal server error" {
				t.Errorf("got body %q, want %q", body, "internal server error")
			}
			if strings.Contains(body, "secret db connection string failed") {
				t.Errorf("response body leaked sensitive internal error details: %s", body)
			}
		})
	}
}
