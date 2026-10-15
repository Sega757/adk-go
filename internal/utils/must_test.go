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

package utils_test

import (
	"errors"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/internal/utils"
)

type mockAgent struct {
	agent.Agent
}

func TestMust(t *testing.T) {
	t.Run("success returns agent", func(t *testing.T) {
		expected := &mockAgent{}
		got := utils.Must(expected, nil)
		if got != expected {
			t.Errorf("Must() = %v, want %v", got, expected)
		}
	})

	t.Run("error panics with error", func(t *testing.T) {
		errTest := errors.New("test error")
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Must() did not panic on non-nil error")
			}
			err, ok := r.(error)
			if !ok || !errors.Is(err, errTest) {
				t.Errorf("Must() panic = %v, want %v", r, errTest)
			}
		}()

		utils.Must[*mockAgent](nil, errTest)
	})
}
