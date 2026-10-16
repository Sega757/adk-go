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
	"time"

	"google.golang.org/adk/v2/server/adkrest/internal/models"
	"google.golang.org/adk/v2/session"
)

type nilStateEventsSession struct{}

func (s *nilStateEventsSession) ID() string             { return "s1" }
func (s *nilStateEventsSession) AppName() string        { return "app1" }
func (s *nilStateEventsSession) UserID() string         { return "u1" }
func (s *nilStateEventsSession) State() session.State   { return nil }
func (s *nilStateEventsSession) Events() session.Events { return nil }
func (s *nilStateEventsSession) LastUpdateTime() time.Time {
	return time.Unix(1000, 0)
}

func TestFromSession_NilGuard(t *testing.T) {
	t.Run("nil_session_returns_error", func(t *testing.T) {
		_, err := models.FromSession(nil)
		if err == nil {
			t.Fatal("expected error when passing nil session, got nil")
		}
	})

	t.Run("nil_state_and_events_handled_safely", func(t *testing.T) {
		sess, err := models.FromSession(&nilStateEventsSession{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sess.ID != "s1" || sess.AppName != "app1" || sess.UserID != "u1" {
			t.Errorf("unexpected session fields: %+v", sess)
		}
		if len(sess.Events) != 0 {
			t.Errorf("expected empty events slice, got %d events", len(sess.Events))
		}
	})
}
