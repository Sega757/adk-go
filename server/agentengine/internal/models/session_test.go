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

package models_test

import (
	"testing"

	"google.golang.org/adk/v2/server/agentengine/internal/models"
)

func TestFromSession_NilSafety(t *testing.T) {
	t.Run("nil session does not panic", func(t *testing.T) {
		got := models.FromSession(nil)
		if got.UserID != "" {
			t.Errorf("expected empty UserID for nil session, got %q", got.UserID)
		}
		if got.State == nil {
			t.Errorf("expected non-nil State map for nil session")
		}
		if got.Events == nil {
			t.Errorf("expected non-nil Events slice for nil session")
		}
	})
}
