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

package helper

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmitJSONError(t *testing.T) {
	rw := httptest.NewRecorder()
	secretErr := errors.New("sensitive database connection string leaked")

	err := EmitJSONError(rw, secretErr)
	if err != nil {
		t.Fatalf("EmitJSONError failed: %v", err)
	}

	body := rw.Body.String()
	if strings.Contains(body, "sensitive database") {
		t.Errorf("EmitJSONError leaked error details in body: %s", body)
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got["error"] != "internal server error" {
		t.Errorf("EmitJSONError response = %q, want %q", got["error"], "internal server error")
	}
}
