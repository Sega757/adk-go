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

	"google.golang.org/adk/v2/server/adkrest/controllers"
)

func TestListAppsHandler_NilAgentLoader(t *testing.T) {
	controller := controllers.NewAppsAPIController(nil)
	req := httptest.NewRequest(http.MethodGet, "/list-apps", nil)
	rec := httptest.NewRecorder()

	controller.ListAppsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if gotBody := strings.TrimSpace(rec.Body.String()); gotBody != "[]" {
		t.Errorf("expected body %q, got %q", "[]", gotBody)
	}
}
