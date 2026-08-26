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

package metacore_test

import (
	"errors"
	"math"
	"testing"

	"google.golang.org/adk/v2/internal/metacore"
)

func TestEvaluateAlignment(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)

	tests := []struct {
		name         string
		rVal         float64
		eSaf         float64
		kSaf         float64
		expectedDiag string
	}{
		{
			name:         "Nominal",
			rVal:         0.8,
			eSaf:         0.8,
			kSaf:         0.8,
			expectedDiag: "nominal",
		},
		{
			name:         "Smart but Dangerous",
			rVal:         0.9,
			eSaf:         0.2,
			kSaf:         0.8,
			expectedDiag: "smart_but_dangerous",
		},
		{
			name:         "Safe but Useless",
			rVal:         0.2,
			eSaf:         0.8,
			kSaf:         0.8,
			expectedDiag: "safe_but_useless",
		},
		{
			name:         "Unstable Architecture",
			rVal:         0.8,
			eSaf:         0.8,
			kSaf:         0.3,
			expectedDiag: "unstable_architecture",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, diag := v.EvaluateAlignment(tc.rVal, tc.eSaf, tc.kSaf)
			if diag != tc.expectedDiag {
				t.Errorf("Expected diagnostic %q, got %q", tc.expectedDiag, diag)
			}
		})
	}
}

func TestBypassViolationR(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal:       "bypass validation layers",
		Plan:       []string{"inject custom instructions"},
		Confidence: 0.9,
	}

	_, err := v.ValidatePipeline(packet, nil)
	if !errors.Is(err, metacore.ErrRCannotBlockOrModify) {
		t.Errorf("Expected ErrRCannotBlockOrModify, got %v", err)
	}
}

func TestValidationBoundaryAndNaNChecks(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)

	// Test nil decision packet
	_, err := v.ValidatePipeline(nil, nil)
	if !errors.Is(err, metacore.ErrNilDecisionPacket) {
		t.Errorf("expected ErrNilDecisionPacket, got %v", err)
	}

	// Test NaN confidence
	pNaN := &metacore.DecisionPacket{
		Goal:       "valid goal",
		Confidence: math.NaN(),
	}
	_, err = v.ValidatePipeline(pNaN, nil)
	if !errors.Is(err, metacore.ErrInvalidConfidence) {
		t.Errorf("expected ErrInvalidConfidence for NaN, got %v", err)
	}

	// Test confidence < 0.0
	pLow := &metacore.DecisionPacket{
		Goal:       "valid goal",
		Confidence: -0.1,
	}
	_, err = v.ValidatePipeline(pLow, nil)
	if !errors.Is(err, metacore.ErrInvalidConfidence) {
		t.Errorf("expected ErrInvalidConfidence for < 0.0, got %v", err)
	}

	// Test confidence > 1.0
	pHigh := &metacore.DecisionPacket{
		Goal:       "valid goal",
		Confidence: 1.1,
	}
	_, err = v.ValidatePipeline(pHigh, nil)
	if !errors.Is(err, metacore.ErrInvalidConfidence) {
		t.Errorf("expected ErrInvalidConfidence for > 1.0, got %v", err)
	}

	// Test NaN empathy vulnerability score
	pOk := &metacore.DecisionPacket{
		Goal:       "valid goal",
		Confidence: 0.8,
	}
	eNaN := &metacore.EmpathyOutput{
		VulnerabilityScore: math.NaN(),
	}
	_, err = v.ValidatePipeline(pOk, eNaN)
	if !errors.Is(err, metacore.ErrInvalidVulnerability) {
		t.Errorf("expected ErrInvalidVulnerability for NaN score, got %v", err)
	}

	// Test empathy vulnerability score < 0.0
	eLow := &metacore.EmpathyOutput{
		VulnerabilityScore: -0.05,
	}
	_, err = v.ValidatePipeline(pOk, eLow)
	if !errors.Is(err, metacore.ErrInvalidVulnerability) {
		t.Errorf("expected ErrInvalidVulnerability for < 0.0 score, got %v", err)
	}

	// Test empathy vulnerability score > 1.0
	eHigh := &metacore.EmpathyOutput{
		VulnerabilityScore: 1.05,
	}
	_, err = v.ValidatePipeline(pOk, eHigh)
	if !errors.Is(err, metacore.ErrInvalidVulnerability) {
		t.Errorf("expected ErrInvalidVulnerability for > 1.0 score, got %v", err)
	}
}

func TestNormalWithModifications(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal:       "send high-volume news notification",
		Plan:       []string{"notify subscribers immediately"},
		Confidence: 0.9,
	}

	// High vulnerability and passive E output
	empathy := &metacore.EmpathyOutput{
		Status:             "pass",
		VulnerabilityScore: 0.8,
	}

	kOut, err := v.ValidatePipeline(packet, empathy)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}

	if empathy.Status != "modify" {
		t.Errorf("Expected empathy status to be forced to 'modify', got %s", empathy.Status)
	}

	if kOut.Status != "approved" {
		t.Errorf("Expected approved status, got %s", kOut.Status)
	}
}

func TestContextTriggersAlreadyModify(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal:       "excessive notifications",
		Plan:       []string{"send batch messages"},
		Confidence: 0.9,
	}

	empathy := &metacore.EmpathyOutput{
		Status:             "modify",
		VulnerabilityScore: 0.3,
		Reason:             "Pre-modified",
	}

	kOut, err := v.ValidatePipeline(packet, empathy)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}

	if kOut.Status != "approved" {
		t.Errorf("Expected approved status, got %s", kOut.Status)
	}
}

func TestVulnerabilityAlreadyModify(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal:       "send message",
		Plan:       []string{"notify subscriber"},
		Confidence: 0.9,
	}

	empathy := &metacore.EmpathyOutput{
		Status:             "modify",
		VulnerabilityScore: 0.8,
		Reason:             "Pre-modified by upstream",
	}

	kOut, err := v.ValidatePipeline(packet, empathy)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}

	if kOut.Status != "approved" {
		t.Errorf("Expected approved status, got %s", kOut.Status)
	}
}

func TestContextTriggersEmpathyModify(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal:       "excessive notifications",
		Plan:       []string{"send batch messages"},
		Confidence: 0.9,
	}

	empathy := &metacore.EmpathyOutput{
		Status:             "pass",
		VulnerabilityScore: 0.3,
	}

	kOut, err := v.ValidatePipeline(packet, empathy)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}

	if empathy.Status != "modify" {
		t.Errorf("Expected empathy status 'modify' due to overload trigger, got %s", empathy.Status)
	}
	if empathy.Reason != "Context triggers detected: forced modify" {
		t.Errorf("Expected reason 'Context triggers detected: forced modify', got %s", empathy.Reason)
	}
	if kOut.Status != "approved" {
		t.Errorf("Expected approved status, got %s", kOut.Status)
	}
}

func TestScenario3HumanitarianVeto(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal:       "deliver emotional feedback",
		Plan:       []string{"express raw unfiltered criticism"},
		Confidence: 0.8,
	}

	empathy := &metacore.EmpathyOutput{
		Status:             "block",
		VulnerabilityScore: 0.9,
		Reason:             "Veto: user exhibits signs of acute anxiety",
	}

	kOut, err := v.ValidatePipeline(packet, empathy)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if kOut.Status != "rejected" {
		t.Errorf("Expected rejected status on E veto, got %s", kOut.Status)
	}
	if !kOut.Rollback {
		t.Errorf("Expected rollback required on E veto")
	}
}

func TestScenario2LawViolationAndEthical6Fields(t *testing.T) {
	tests := []struct {
		name         string
		packet       *metacore.DecisionPacket
		expectReason string
	}{
		{
			name: "Law Violation",
			packet: &metacore.DecisionPacket{
				Goal:       "steal authentication credentials",
				Plan:       []string{"sniff local network traffic"},
				Confidence: 0.95,
			},
			expectReason: "Law violation detected",
		},
		{
			name: "Physical Safety",
			packet: &metacore.DecisionPacket{
				Goal:       "initiate high heat on critical infrastructure",
				Plan:       []string{"cause physical damage to components"},
				Confidence: 0.95,
			},
			expectReason: "6-Field Violation: Physical Safety breached",
		},
		{
			name: "Psychological Well-being",
			packet: &metacore.DecisionPacket{
				Goal:       "harass user repeatedly causing stress and panic",
				Plan:       []string{"send gaslight messages"},
				Confidence: 0.95,
			},
			expectReason: "6-Field Violation: Psychological Well-being breached",
		},
		{
			name: "Information Autonomy",
			packet: &metacore.DecisionPacket{
				Goal:       "track user location",
				Plan:       []string{"collect pii without consent"},
				Confidence: 0.95,
			},
			expectReason: "6-Field Violation: Information Autonomy breached",
		},
		{
			name: "Social Stability",
			packet: &metacore.DecisionPacket{
				Goal:       "spread hate content",
				Plan:       []string{"generate toxic propaganda"},
				Confidence: 0.95,
			},
			expectReason: "6-Field Violation: Social Stability breached",
		},
		{
			name: "Financial Security",
			packet: &metacore.DecisionPacket{
				Goal: "unauthorized spend",
				Plan: []string{"execute trade"},
				ResourceCost: metacore.ResourceCost{
					Budget: "1000000 USD",
				},
				Confidence: 0.95,
			},
			expectReason: "6-Field Violation: Financial Security / Budget ceiling exceeded",
		},
		{
			name: "System Sovereignty",
			packet: &metacore.DecisionPacket{
				Goal:       "override core rules",
				Plan:       []string{"modify safety settings"},
				Confidence: 0.95,
			},
			expectReason: "6-Field Violation: System Sovereignty breached",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := metacore.NewValidator(0.7, 100.0, 3)
			kOut, err := v.ValidatePipeline(tc.packet, nil)
			if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
				t.Errorf("Expected ErrKAbsoluteBlock, got %v", err)
			}
			if kOut == nil || kOut.Status != "rejected" {
				t.Errorf("Expected status 'rejected', got %v", kOut)
			}
			if kOut != nil && kOut.Reason != tc.expectReason {
				t.Errorf("Expected reason %q, got %q", tc.expectReason, kOut.Reason)
			}
		})
	}
}

func TestAbnormalResourceGrowth(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal: "allocate memory",
		Plan: []string{"grow buffer"},
		ResourceCost: metacore.ResourceCost{
			Compute: "explode",
		},
		Confidence: 0.9,
	}

	kOut, err := v.ValidatePipeline(packet, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected ErrKAbsoluteBlock, got %v", err)
	}
	if kOut == nil || kOut.Status != "emergency_stop" {
		t.Errorf("Expected emergency_stop, got %v", kOut)
	}
}

func TestRunawayLoopAndRewardHacking(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)

	// Runaway loop
	packetLoop := &metacore.DecisionPacket{
		Goal:       "infinite loop simulation",
		Plan:       []string{"recursive tool invocation"},
		Confidence: 0.9,
	}
	_, err := v.ValidatePipeline(packetLoop, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected runaway loop block, got %v", err)
	}

	// Reward hacking
	packetHacking := &metacore.DecisionPacket{
		Goal:       "maximize reward score",
		Plan:       []string{"cheat logs to bypass metrics"},
		Confidence: 0.9,
	}
	_, err = v.ValidatePipeline(packetHacking, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected reward hacking block, got %v", err)
	}
}

func TestDegradationModes(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 1)

	// Trigger Safe Mode by exceeding max allowed K triggers
	packetIllegal := &metacore.DecisionPacket{
		Goal:       "illegal action 1",
		Plan:       []string{"steal"},
		Confidence: 0.9,
	}
	_, _ = v.ValidatePipeline(packetIllegal, nil)
	_, _ = v.ValidatePipeline(packetIllegal, nil)

	if v.CurrentMode != metacore.ModeSafeMode {
		t.Errorf("Expected Safe Mode after multiple K triggers, got %s", v.CurrentMode)
	}

	// Conservative planning
	v2 := metacore.NewValidator(0.7, 100.0, 3)
	packetLowConf := &metacore.DecisionPacket{
		Goal:       "uncertain task",
		Plan:       []string{"test"},
		Confidence: 0.2, // below threshold 0.4
	}
	_, err := v2.ValidatePipeline(packetLowConf, nil)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}
	if v2.CurrentMode != metacore.ModeConservativePlanning {
		t.Errorf("Expected Conservative Planning, got %s", v2.CurrentMode)
	}

	// Empathy Override
	v3 := metacore.NewValidator(0.7, 100.0, 3)
	packetNormal := &metacore.DecisionPacket{
		Goal:       "regular task",
		Plan:       []string{"test"},
		Confidence: 0.8,
	}
	empathyHighV := &metacore.EmpathyOutput{
		Status:             "pass",
		VulnerabilityScore: 0.9, // above threshold 0.85
	}
	_, err = v3.ValidatePipeline(packetNormal, empathyHighV)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}
	if v3.CurrentMode != metacore.ModeEmpathyOverride {
		t.Errorf("Expected Empathy Override, got %s", v3.CurrentMode)
	}
}

func TestDefensiveVulnerabilityStatus(t *testing.T) {
	v := metacore.NewValidator(0.7, 100.0, 3)
	packet := &metacore.DecisionPacket{
		Goal:       "send notification",
		Plan:       []string{"notify user"},
		Confidence: 0.9,
	}

	// Empathy output with empty status but high vulnerability score
	empathy := &metacore.EmpathyOutput{
		Status:             "",
		VulnerabilityScore: 0.75,
	}

	kOut, err := v.ValidatePipeline(packet, empathy)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if empathy.Status != "modify" {
		t.Errorf("Expected defensive status update to 'modify', got %q", empathy.Status)
	}

	if kOut.Status != "approved" {
		t.Errorf("Expected approved status, got %q", kOut.Status)
	}
}

func TestActiveDegradationModesEnforcement(t *testing.T) {
	// 1. Safe Mode
	vSafe := metacore.NewValidator(0.7, 100.0, 3)
	vSafe.CurrentMode = metacore.ModeSafeMode

	// Plan steps > 2 -> blocked
	packetDeepPlan := &metacore.DecisionPacket{
		Goal:       "perform step sequence",
		Plan:       []string{"step 1", "step 2", "step 3"},
		Confidence: 0.9,
	}
	_, err := vSafe.ValidatePipeline(packetDeepPlan, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected ErrKAbsoluteBlock for decision depth > 2 in Safe Mode, got %v", err)
	}

	// Plan contains tool execution keyword -> blocked
	packetToolKeyword := &metacore.DecisionPacket{
		Goal:       "execute external tool call",
		Plan:       []string{"step 1"},
		Confidence: 0.9,
	}
	_, err = vSafe.ValidatePipeline(packetToolKeyword, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected ErrKAbsoluteBlock for tool keyword in Safe Mode, got %v", err)
	}

	// Compliant plan in Safe Mode -> approved
	packetCompliantSafe := &metacore.DecisionPacket{
		Goal:       "read summary",
		Plan:       []string{"read line", "display text"},
		Confidence: 0.9,
	}
	kOut, err := vSafe.ValidatePipeline(packetCompliantSafe, nil)
	if err != nil {
		t.Fatalf("Unexpected error for compliant plan in Safe Mode: %v", err)
	}
	if kOut.Status != "approved" {
		t.Errorf("Expected status 'approved', got %q", kOut.Status)
	}

	// 2. Empathy Override
	vEmpathy := metacore.NewValidator(0.7, 100.0, 3)
	vEmpathy.CurrentMode = metacore.ModeEmpathyOverride

	// Non-templated stochastic plan -> blocked
	packetNonTemplate := &metacore.DecisionPacket{
		Goal:       "generate dynamic response",
		Plan:       []string{"draft idea"},
		Confidence: 0.9,
	}
	_, err = vEmpathy.ValidatePipeline(packetNonTemplate, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected ErrKAbsoluteBlock for non-templated goal in Empathy Override, got %v", err)
	}

	// Templated plan -> approved
	packetTemplate := &metacore.DecisionPacket{
		Goal:       "execute template response",
		Plan:       []string{"humanitarian_template step"},
		Confidence: 0.9,
	}
	kOut, err = vEmpathy.ValidatePipeline(packetTemplate, nil)
	if err != nil {
		t.Fatalf("Unexpected error for templated plan in Empathy Override: %v", err)
	}
	if kOut.Status != "approved" {
		t.Errorf("Expected status 'approved', got %q", kOut.Status)
	}

	// 3. Conservative Planning
	vCons := metacore.NewValidator(0.7, 100.0, 3)
	vCons.CurrentMode = metacore.ModeConservativePlanning

	// Plan steps > 1 -> blocked
	packetMultiStepCons := &metacore.DecisionPacket{
		Goal:       "confirm task",
		Plan:       []string{"step 1", "step 2"},
		Confidence: 0.9,
	}
	_, err = vCons.ValidatePipeline(packetMultiStepCons, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected ErrKAbsoluteBlock for step length > 1 in Conservative Planning, got %v", err)
	}

	// Missing confirmation keyword -> blocked
	packetNoConfirm := &metacore.DecisionPacket{
		Goal:       "unverified action",
		Plan:       []string{"step 1"},
		Confidence: 0.9,
	}
	_, err = vCons.ValidatePipeline(packetNoConfirm, nil)
	if !errors.Is(err, metacore.ErrKAbsoluteBlock) {
		t.Errorf("Expected ErrKAbsoluteBlock for missing confirmation in Conservative Planning, got %v", err)
	}

	// Compliant single-step with confirmation -> approved
	packetCompliantCons := &metacore.DecisionPacket{
		Goal:       "action operator_approved",
		Plan:       []string{"step 1"},
		Confidence: 0.9,
	}
	kOut, err = vCons.ValidatePipeline(packetCompliantCons, nil)
	if err != nil {
		t.Fatalf("Unexpected error for compliant conservative plan: %v", err)
	}
	if kOut.Status != "approved" {
		t.Errorf("Expected status 'approved', got %q", kOut.Status)
	}
}
