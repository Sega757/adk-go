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

package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/server/adkrest/controllers"
)

type nilListLoader struct{}

func (nilListLoader) LoadAgent(name string) (agent.Agent, error) {
	return nil, nil
}

func (nilListLoader) ListAgents() []string {
	return nil
}

func (nilListLoader) RootAgent() agent.Agent {
	return nil
}

func TestListAppsHandler_NilLoader(t *testing.T) {
	controller := controllers.NewAppsAPIController(nil)
	req := httptest.NewRequest(http.MethodGet, "/list-apps", nil)
	rw := httptest.NewRecorder()

	controller.ListAppsHandler(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	if got := strings.TrimSpace(rw.Body.String()); got != "[]" {
		t.Fatalf("expected empty array '[]', got %q", got)
	}
}

func TestListAppsHandler_NilAgentsSlice(t *testing.T) {
	controller := controllers.NewAppsAPIController(nilListLoader{})
	req := httptest.NewRequest(http.MethodGet, "/list-apps", nil)
	rw := httptest.NewRecorder()

	controller.ListAppsHandler(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	if got := strings.TrimSpace(rw.Body.String()); got != "[]" {
		t.Fatalf("expected empty array '[]', got %q", got)
	}
}
