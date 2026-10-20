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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/server/adkrest/controllers"
)

type nilListAgentLoader struct{}

func (n *nilListAgentLoader) ListAgents() []string {
	return nil
}

func (n *nilListAgentLoader) LoadAgent(name string) (agent.Agent, error) {
	return nil, nil
}

func (n *nilListAgentLoader) RootAgent() agent.Agent {
	return nil
}

func TestListAppsHandler(t *testing.T) {
	t.Run("nil_agent_loader", func(t *testing.T) {
		controller := controllers.NewAppsAPIController(nil)
		req := httptest.NewRequest(http.MethodGet, "/apps", nil)
		rec := httptest.NewRecorder()

		controller.ListAppsHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
		}
		if gotBody := strings.TrimSpace(rec.Body.String()); gotBody != "[]" {
			t.Errorf("got body %q, want %q", gotBody, "[]")
		}
	})

	t.Run("agent_loader_returns_nil_slice", func(t *testing.T) {
		controller := controllers.NewAppsAPIController(&nilListAgentLoader{})
		req := httptest.NewRequest(http.MethodGet, "/apps", nil)
		rec := httptest.NewRecorder()

		controller.ListAppsHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
		}
		if gotBody := strings.TrimSpace(rec.Body.String()); gotBody != "[]" {
			t.Errorf("got body %q, want %q", gotBody, "[]")
		}
	})

	t.Run("agent_loader_with_agents", func(t *testing.T) {
		dummyAgent, err := agent.New(agent.Config{Name: "test_agent"})
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		loader := agent.NewSingleLoader(dummyAgent)
		controller := controllers.NewAppsAPIController(loader)
		req := httptest.NewRequest(http.MethodGet, "/apps", nil)
		rec := httptest.NewRecorder()

		controller.ListAppsHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
		}
		if gotBody := strings.TrimSpace(rec.Body.String()); gotBody != `["test_agent"]` {
			t.Errorf("got body %q, want %q", gotBody, `["test_agent"]`)
		}
	})
}
