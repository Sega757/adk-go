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

package sessioninternal

import (
	"sync"
	"time"

	"google.golang.org/adk/v2/session"
)

// Session represents an in-memory session data structure shared across session implementations.
type Session struct {
	AppName   string
	UserID    string
	SessionID string

	// Mu guards all mutable fields.
	Mu        sync.RWMutex
	Events    []*session.Event
	State     map[string]any
	UpdatedAt time.Time
}
