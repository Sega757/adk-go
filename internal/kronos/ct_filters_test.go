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

package kronos_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/adk/internal/kronos"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		ip        string
		compliant bool
	}{
		// Non-compliant private ranges
		{"127.0.0.1", false},
		{"10.20.30.40", false},
		{"192.168.1.1", false},
		{"169.254.169.254", false},
		{"172.16.0.5", false},
		// Compliant public ranges
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"142.250.190.46", true},
		// Invalid formats
		{"not-an-ip", false},
		{"999.999.999.999", false},
	}

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			got := kronos.ValidateIP(tc.ip)
			if got != tc.compliant {
				t.Errorf("ValidateIP(%q) = %v; want %v", tc.ip, got, tc.compliant)
			}
		})
	}
}

func TestFilterAudit(t *testing.T) {
	// Standard validation
	report, err := kronos.FilterAudit("South Africa vs Mexico", 1.85, "Wilton Sampaio", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Event != "South Africa vs Mexico" {
		t.Errorf("expected Event to be 'South Africa vs Mexico', got %q", report.Event)
	}

	if !strings.Contains(report.ProtocolCT.Logic, "1.85") {
		t.Errorf("expected Logic protocol to include xG divergence 1.85, got %q", report.ProtocolCT.Logic)
	}

	// Delta-Chi refuse execution check
	_, err = kronos.FilterAudit("South Africa vs Mexico", 1.85, "Wilton Sampaio", true)
	if err == nil {
		t.Fatal("expected error under Delta-Chi refuse execution conditions, got nil")
	}
}

func TestValidateIPHandler(t *testing.T) {
	h := kronos.NewHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Post compliant IP
	bodyBytes, _ := json.Marshal(kronos.IPRequest{IP: "8.8.8.8"})
	req := httptest.NewRequest(http.MethodPost, "/api/validate-ip", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rec.Code)
	}

	var resp kronos.IPResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Compliant {
		t.Errorf("expected IP 8.8.8.8 to be compliant, got non-compliant")
	}

	// Post loopback IP
	bodyBytes, _ = json.Marshal(kronos.IPRequest{IP: "127.0.0.1"})
	req = httptest.NewRequest(http.MethodPost, "/api/validate-ip", bytes.NewReader(bodyBytes))
	rec = httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rec.Code)
	}

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Compliant {
		t.Errorf("expected IP 127.0.0.1 to be non-compliant, got compliant")
	}
}

func TestExecuteAuditHandler(t *testing.T) {
	h := kronos.NewHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Standard audit
	bodyBytes, _ := json.Marshal(kronos.AuditRequest{
		Event:         "JS Kabylie vs MC Alger",
		XGAnomaly:     1.17,
		RefereeName:   "Wilton Sampaio",
		TriggerDeltaX: false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/execute-audit", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rec.Code)
	}

	var report kronos.AuditReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if report.Event != "JS Kabylie vs MC Alger" {
		t.Errorf("expected event 'JS Kabylie vs MC Alger', got %q", report.Event)
	}

	// Delta-Chi Refusal execution
	bodyBytes, _ = json.Marshal(kronos.AuditRequest{
		Event:         "South Africa vs Mexico",
		XGAnomaly:     1.85,
		RefereeName:   "Wilton Sampaio",
		TriggerDeltaX: true,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/execute-audit", bytes.NewReader(bodyBytes))
	rec = httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected StatusForbidden, got %v", rec.Code)
	}

	var errResp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if errResp["code"] != "DELTA_CHI_TRIGGERED" {
		t.Errorf("expected DELTA_CHI_TRIGGERED, got %q", errResp["code"])
	}
}
