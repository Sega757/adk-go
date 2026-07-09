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

package kronos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// IPRequest represents the payload to check IP address compliance.
type IPRequest struct {
	IP string `json:"ip"`
}

// IPResponse represents the outcome of an IP address validation check.
type IPResponse struct {
	IP       string `json:"ip"`
	Compliant bool   `json:"compliant"`
	Message   string `json:"message"`
}

// AuditRequest represents inputs to trigger the Conjunctive Transparency (C-T) deep audit.
type AuditRequest struct {
	Event         string  `json:"event"`
	XGAnomaly     float64 `json:"xg_anomaly"`
	RefereeName   string  `json:"referee_name"`
	TriggerDeltaX bool    `json:"trigger_deltax"`
}

// Handler handles backend requests for the C-T Amethyst Kernel platform.
type Handler struct{}

// NewHandler constructs a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes configures routing on the supplied ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/validate-ip", h.HandleValidateIP)
	mux.HandleFunc("/api/execute-audit", h.HandleExecuteAudit)
	mux.HandleFunc("/api/health", h.HandleHealth)
}

// HandleValidateIP inspects an incoming IP address for private/restricted space compliance.
func (h *Handler) HandleValidateIP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req IPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	compliant := ValidateIP(req.IP)
	resp := IPResponse{
		IP:        req.IP,
		Compliant: compliant,
	}
	if compliant {
		resp.Message = "IP address is compliant: Outside of local, loopback, private, or reserved subnets."
	} else {
		resp.Message = "IP address is NON-compliant: Belongs to private, loopback, link-local, or reserved address block."
	}

	json.NewEncoder(w).Encode(resp)
}

// HandleExecuteAudit triggers a simulated Conjunctive Transparency 6-field validation.
func (h *Handler) HandleExecuteAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Set defaults for simple GET or blank tests
		req.Event = "South Africa vs Mexico"
		req.XGAnomaly = 1.41
		req.RefereeName = "Wilton Sampaio"
	}

	if req.Event == "" {
		req.Event = "South Africa vs Mexico"
	}
	if req.RefereeName == "" {
		req.RefereeName = "Wilton Sampaio"
	}

	report, err := FilterAudit(req.Event, req.XGAnomaly, req.RefereeName, req.TriggerDeltaX)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   err.Error(),
			"status":  "Technical Amber",
			"code":    "DELTA_CHI_TRIGGERED",
			"message": "Execution halted immediately under Delta-Chi Right of Refusal.",
		})
		return
	}

	json.NewEncoder(w).Encode(report)
}

// HandleHealth serves platform state.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":      "OPERATIONAL",
		"firewall":    "SECURE",
		"amethyst":    "NOMINAL",
		"temperature": "Technical Amber",
		"agent":       "KRONOS: OCTOPUS (V 1.0)",
	})
}
