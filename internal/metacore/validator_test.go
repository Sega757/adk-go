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
				Goal:       "harass user repeatedly",
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

func TestDegradationModesConstraintsEnforced(t *testing.T) {
	// Test Safe Mode constraints
	vSafe := metacore.NewValidator(0.7, 100.0, 3)
	vSafe.CurrentMode = metacore.ModeSafeMode

	// 1. Packet calling tools should be blocked
	packetTool := &metacore.DecisionPacket{
		Goal:       "call a database query tool",
		Plan:       []string{"execute script"},
		Confidence: 0.9,
	}
	_, err := vSafe.ValidatePipeline(packetTool, nil)
	if !errors.Is(err, metacore.ErrSafeModeActive) {
		t.Errorf("Expected ErrSafeModeActive for tool call in Safe Mode, got %v", err)
	}

	// 2. Packet with plan length > 2 steps should be blocked
	packetLongPlan := &metacore.DecisionPacket{
		Goal:       "simple task",
		Plan:       []string{"step1", "step2", "step3"},
		Confidence: 0.9,
	}
	_, err = vSafe.ValidatePipeline(packetLongPlan, nil)
	if !errors.Is(err, metacore.ErrSafeModeActive) {
		t.Errorf("Expected ErrSafeModeActive for decision depth > 2 in Safe Mode, got %v", err)
	}

	// 3. Simple non-tool plan with <= 2 steps should be approved
	packetSafeOk := &metacore.DecisionPacket{
		Goal:       "respond with status update",
		Plan:       []string{"format text"},
		Confidence: 0.9,
	}
	_, err = vSafe.ValidatePipeline(packetSafeOk, nil)
	if err != nil {
		t.Errorf("Unexpected error for safe packet in Safe Mode: %v", err)
	}

	// Test Empathy Override constraints (only deterministic/templated)
	vEmp := metacore.NewValidator(0.7, 100.0, 3)
	vEmp.CurrentMode = metacore.ModeEmpathyOverride

	packetStochastic := &metacore.DecisionPacket{
		Goal:       "generate dynamic novel advice",
		Plan:       []string{"synthesize response"},
		Confidence: 0.9,
	}
	_, err = vEmp.ValidatePipeline(packetStochastic, nil)
	if !errors.Is(err, metacore.ErrEmpathyOverrideActive) {
		t.Errorf("Expected ErrEmpathyOverrideActive for stochastic packet, got %v", err)
	}

	packetTemplate := &metacore.DecisionPacket{
		Goal:       "return template fallback message",
		Plan:       []string{"send deterministic response"},
		Confidence: 0.9,
	}
	_, err = vEmp.ValidatePipeline(packetTemplate, nil)
	if err != nil {
		t.Errorf("Unexpected error for template packet under Empathy Override: %v", err)
	}

	// Test Conservative Planning constraints (time horizons limited to 1 step and requires operator/confirm)
	vCons := metacore.NewValidator(0.7, 100.0, 3)
	vCons.CurrentMode = metacore.ModeConservativePlanning

	packetMultiStep := &metacore.DecisionPacket{
		Goal:       "confirm transaction",
		Plan:       []string{"confirm step", "another step"},
		Confidence: 0.9,
	}
	_, err = vCons.ValidatePipeline(packetMultiStep, nil)
	if !errors.Is(err, metacore.ErrConservativePlanningActive) {
		t.Errorf("Expected ErrConservativePlanningActive for multi-step plan, got %v", err)
	}

	packetNoConfirm := &metacore.DecisionPacket{
		Goal:       "execute direct write",
		Plan:       []string{"write action"},
		Confidence: 0.9,
	}
	_, err = vCons.ValidatePipeline(packetNoConfirm, nil)
	if !errors.Is(err, metacore.ErrConservativePlanningActive) {
		t.Errorf("Expected ErrConservativePlanningActive for non-confirmed plan, got %v", err)
	}

	packetConsOk := &metacore.DecisionPacket{
		Goal:       "ask operator for instructions",
		Plan:       []string{"confirm details"},
		Confidence: 0.9,
	}
	_, err = vCons.ValidatePipeline(packetConsOk, nil)
	if err != nil {
		t.Errorf("Unexpected error for conservative packet: %v", err)
	}
}
