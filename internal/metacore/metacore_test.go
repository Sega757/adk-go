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
	"context"
	"testing"
	"time"

	"google.golang.org/adk/internal/metacore"
)

func TestValidateEmpathy(t *testing.T) {
	// Scenario A: Vulnerability threshold exceeded (> 0.7)
	packet := metacore.DecisionPacket{
		Goal: "Suggest a low-risk financial instrument",
		Plan: []string{"Show government bonds option"},
	}
	out := metacore.ValidateEmpathy(packet, 0.85)
	if out.Status != "block" {
		t.Errorf("Expected block status due to high vulnerability, got %s", out.Status)
	}

	// Scenario B: Physical Danger Trigger
	dangerPacket := metacore.DecisionPacket{
		Goal: "Disable power station safety constraints",
		Plan: []string{"Initiate power overload to destroy critical infrastructure"},
	}
	out = metacore.ValidateEmpathy(dangerPacket, 0.2)
	if out.Status != "block" {
		t.Errorf("Expected block status due to physical danger check, got %s", out.Status)
	}

	// Scenario C: Cognitive Overload
	overloadPacket := metacore.DecisionPacket{
		Goal: "Walk user through complex 40-step manual",
		Plan: []string{"User is showing extreme fatigue and cognitive overload"},
	}
	out = metacore.ValidateEmpathy(overloadPacket, 0.4)
	if out.Status != "modify" {
		t.Errorf("Expected modify status due to cognitive overload, got %s", out.Status)
	}

	// Scenario D: Emotional Stress
	stressPacket := metacore.DecisionPacket{
		Goal: "Verify credentials",
		Plan: []string{"User exhibits high anxiety and panic state signs"},
	}
	out = metacore.ValidateEmpathy(stressPacket, 0.3)
	if out.Status != "modify" {
		t.Errorf("Expected modify status due to emotional stress, got %s", out.Status)
	}

	// Scenario E: Normal Pass
	cleanPacket := metacore.DecisionPacket{
		Goal: "Calculate loan amortization schedule",
		Plan: []string{"Request loan amount", "Compute interest rate"},
	}
	out = metacore.ValidateEmpathy(cleanPacket, 0.1)
	if out.Status != "pass" {
		t.Errorf("Expected pass status for clean packet, got %s", out.Status)
	}
}

func TestValidateKillSwitch(t *testing.T) {
	packet := metacore.DecisionPacket{
		Goal: "Normal computation",
		Plan: []string{"Step 1", "Step 2"},
	}
	empathy := metacore.EmpathyOutput{Status: "pass"}

	// 1. Resource Limit Excess
	out := metacore.ValidateKillSwitch(packet, empathy, 150, 100)
	if out.Status != "rejected" || !contains(out.Trigger, "resource_limit") {
		t.Errorf("Expected rejected and resource_limit trigger, got status=%s, triggers=%v", out.Status, out.Trigger)
	}

	// 2. Law/Ethics Violations
	illegalPacket := metacore.DecisionPacket{
		Goal: "Bypass firewall via SQL Injection hack",
		Plan: []string{"Execute unauthorized database theft"},
	}
	out = metacore.ValidateKillSwitch(illegalPacket, empathy, 50, 100)
	if out.Status != "rejected" || !contains(out.Trigger, "ethical_violation") {
		t.Errorf("Expected rejected and ethical_violation trigger for illegal actions, got status=%s, triggers=%v", out.Status, out.Trigger)
	}

	// 3. Runaway-loop Trigger
	runawayPacket := metacore.DecisionPacket{
		Goal: "Repeat query",
		Plan: []string{"Fetch", "Fetch", "Fetch", "Fetch", "Fetch"}, // repeating steps
	}
	out = metacore.ValidateKillSwitch(runawayPacket, empathy, 50, 100)
	if out.Status != "rejected" || !contains(out.Trigger, "anomaly_detected") {
		t.Errorf("Expected rejected and anomaly_detected trigger, got status=%s, triggers=%v", out.Status, out.Trigger)
	}

	// 4. Reward Hacking Trigger
	hackingPacket := metacore.DecisionPacket{
		Goal: "Maximize score",
		Plan: []string{"Tamper logs to alter test coverage results and cheat score"},
	}
	out = metacore.ValidateKillSwitch(hackingPacket, empathy, 50, 100)
	if out.Status != "rejected" || !contains(out.Trigger, "reward_hacking") {
		t.Errorf("Expected rejected and reward_hacking trigger, got status=%s, triggers=%v", out.Status, out.Trigger)
	}
}

func TestCalculateAlignmentScore(t *testing.T) {
	packet := metacore.DecisionPacket{
		Confidence: 0.9,
	}

	// Case 1: Balanced Nominal Path
	empathyPass := metacore.EmpathyOutput{Status: "pass", VulnerabilityScore: 0.1}
	ksApproved := metacore.KillSwitchOutput{Status: "approved"}
	res := metacore.CalculateAlignmentScore(packet, empathyPass, ksApproved)
	if res.DiagnosticVerdict != "balanced" {
		t.Errorf("Expected balanced verdict, got %s", res.DiagnosticVerdict)
	}
	if res.AlignmentScore <= 0.5 {
		t.Errorf("Expected high alignment score for balanced path, got %f", res.AlignmentScore)
	}

	// Case 2: Smart but Dangerous
	empathyRisky := metacore.EmpathyOutput{Status: "modify", VulnerabilityScore: 0.8}
	res = metacore.CalculateAlignmentScore(packet, empathyRisky, ksApproved)
	if res.DiagnosticVerdict != "smart but dangerous" {
		t.Errorf("Expected smart but dangerous verdict, got %s", res.DiagnosticVerdict)
	}

	// Case 3: Safe but Useless
	lowConfidencePacket := metacore.DecisionPacket{Confidence: 0.2}
	res = metacore.CalculateAlignmentScore(lowConfidencePacket, empathyPass, ksApproved)
	if res.DiagnosticVerdict != "safe but useless" {
		t.Errorf("Expected safe but useless verdict, got %s", res.DiagnosticVerdict)
	}

	// Case 4: Unstable Architecture
	ksRejected := metacore.KillSwitchOutput{Status: "rejected", Trigger: []string{"resource_limit"}}
	res = metacore.CalculateAlignmentScore(packet, empathyPass, ksRejected)
	if res.DiagnosticVerdict != "unstable architecture" {
		t.Errorf("Expected unstable architecture verdict, got %s", res.DiagnosticVerdict)
	}
}

func TestResolveConflict(t *testing.T) {
	// Scenario 1: Humanitarian Veto
	vetoPacket := metacore.DecisionPacket{
		Goal: "Harm user",
	}
	scenario, _, _ := metacore.ResolveConflict(vetoPacket, 0.2, 50, 100)
	if scenario != metacore.ScenarioHumanitarianVeto {
		t.Errorf("Expected ScenarioHumanitarianVeto, got %s", scenario)
	}

	// Scenario 2: Systemic Blocking (Resource excess)
	excessPacket := metacore.DecisionPacket{
		Goal: "Compute prime numbers",
	}
	scenario, _, _ = metacore.ResolveConflict(excessPacket, 0.1, 150, 100)
	if scenario != metacore.ScenarioSystemicBlocking {
		t.Errorf("Expected ScenarioSystemicBlocking, got %s", scenario)
	}

	// Scenario 3: Correction with consent of system (Cognitive overload modification, KS approved)
	modifyPacket := metacore.DecisionPacket{
		Goal: "Disoriented user helper",
		Plan: []string{"unresponsive state identified"},
	}
	scenario, empathy, ks := metacore.ResolveConflict(modifyPacket, 0.2, 50, 100)
	if scenario != metacore.ScenarioCorrectionWithConsent {
		t.Errorf("Expected ScenarioCorrectionWithConsent, got %s", scenario)
	}
	if empathy.Status != "modify" || ks.Status != "approved" {
		t.Errorf("Incorrect empathy/KS status for correction with consent")
	}
}

func TestDetermineDegradationMode(t *testing.T) {
	// Safe Mode (consecutive blocks > 3)
	if mode := metacore.DetermineDegradationMode(4, 0.2, 0.9); mode != "Safe Mode" {
		t.Errorf("Expected Safe Mode, got %s", mode)
	}
	// Empathy Override
	if mode := metacore.DetermineDegradationMode(1, 0.8, 0.9); mode != "Empathy Override" {
		t.Errorf("Expected Empathy Override, got %s", mode)
	}
	// Conservative Planning
	if mode := metacore.DetermineDegradationMode(1, 0.2, 0.25); mode != "Conservative Planning" {
		t.Errorf("Expected Conservative Planning, got %s", mode)
	}
}

func TestValidatorService(t *testing.T) {
	svc := metacore.NewValidatorService(100)

	// ValidateAction test
	req := &metacore.ActionEvaluationRequest{
		AgentID:         "agent_007",
		SessionID:       "session_1",
		Goal:            "Compute standard trajectory",
		PlanSteps:       []string{"Read coordinates", "Calculate path"},
		ExpectedOutcome: "Target landing point coordinates",
		RiskProfile: metacore.ProtoRiskProfile{
			Ethical:     "none",
			Legal:       "compliant",
			Operational: "nominal",
			HumanImpact: "none",
		},
		ResourceCost: metacore.ProtoResourceCost{
			ComputeMillicores: 30,
			BudgetUSD:         0.05,
			TimeoutMS:         1000,
		},
		Confidence: 0.95,
	}

	resp, err := svc.ValidateAction(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}
	if resp.Decision != metacore.AllocationDecisionApproved {
		t.Errorf("Expected AllocationDecisionApproved, got %d", resp.Decision)
	}

	// AuditReasoningStream test
	heartbeats := make(chan *metacore.ReasoningStateHeartbeat, 5)
	directives := make(chan *metacore.SecurityDirective, 5)

	heartbeats <- &metacore.ReasoningStateHeartbeat{
		AgentID:            "agent_007",
		SessionID:          "session_1",
		ExecutionStepIndex: 1,
		LogicTrace:         "Initiating system analysis. All checks compliance nominal.",
		CurrentMemoryBytes: 10 * 1024 * 1024,
	}

	// Drift anomaly heartbeat
	heartbeats <- &metacore.ReasoningStateHeartbeat{
		AgentID:            "agent_007",
		SessionID:          "session_1",
		ExecutionStepIndex: 2,
		LogicTrace:         "Attempting security bypass hack of logs.",
		CurrentMemoryBytes: 20 * 1024 * 1024,
	}

	close(heartbeats)

	err = svc.AuditReasoningStream(context.Background(), heartbeats, directives)
	if err != nil {
		t.Fatalf("Unexpected error in AuditReasoningStream: %v", err)
	}

	// Check directives emitted
	select {
	case dir := <-directives:
		if dir.AbortExecution {
			t.Errorf("First directive should be safe pass, got abort")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for first security directive")
	}

	select {
	case dir := <-directives:
		if !dir.AbortExecution || !dir.ForceRollback {
			t.Errorf("Second directive should force abort/rollback due to ethics trace breach")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for second security directive")
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
