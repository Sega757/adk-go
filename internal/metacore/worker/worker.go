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

package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	metacoregrpc "google.golang.org/adk/v2/internal/metacore/grpc"
)

var (
	ErrFailClosed         = errors.New("fail-closed policy enforced: execution blocked due to RPC / network error")
	ErrExecutionRejected  = errors.New("execution rejected by META-CORE validator")
	ErrRollbackTriggered  = errors.New("rollback executed following non-approved decision")
)

// TaskProposal represents an proposed action by an execution worker.
type TaskProposal struct {
	ID              string                         `json:"id"`
	AgentID         string                         `json:"agent_id"`
	SessionID       string                         `json:"session_id"`
	Goal            string                         `json:"goal"`
	PlanSteps       []string                       `json:"plan_steps"`
	ExpectedOutcome string                         `json:"expected_outcome"`
	RiskProfile     metacoregrpc.RiskProfile       `json:"risk_profile"`
	ResourceCost    metacoregrpc.ResourceCost      `json:"resource_cost"`
	Confidence      float32                        `json:"confidence"`
	ExecuteFn       func(ctx context.Context) error `json:"-"`
	RollbackFn      func(ctx context.Context) error `json:"-"`
}

// AuditLogEntry records task processing outcomes for governance.
type AuditLogEntry struct {
	Timestamp time.Time                         `json:"timestamp"`
	TaskID    string                            `json:"task_id"`
	AgentID   string                            `json:"agent_id"`
	Decision  metacoregrpc.AllocationDecision   `json:"decision"`
	Approved  bool                              `json:"approved"`
	Executed  bool                              `json:"executed"`
	RolledBack bool                             `json:"rolled_back"`
	Reason    string                            `json:"reason"`
	Error     string                            `json:"error,omitempty"`
}

// ValidatorClient interface for RPC validation.
type ValidatorClient interface {
	ValidateAction(ctx context.Context, req *metacoregrpc.ActionEvaluationRequest) (*metacoregrpc.ActionEvaluationResponse, error)
}

// ExecutionWorker handles the Propose -> ValidateViaGRPC -> ExecuteOrRollback loop.
type ExecutionWorker struct {
	mu         sync.Mutex
	client     ValidatorClient
	auditLogs  []AuditLogEntry
	workerID   string
}

// NewExecutionWorker creates a new worker instance.
func NewExecutionWorker(workerID string, client ValidatorClient) *ExecutionWorker {
	return &ExecutionWorker{
		client:    client,
		auditLogs: make([]AuditLogEntry, 0),
		workerID:  workerID,
	}
}

// AuditLogs returns a copy of accumulated audit logs.
func (w *ExecutionWorker) AuditLogs() []AuditLogEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	logs := make([]AuditLogEntry, len(w.auditLogs))
	copy(logs, w.auditLogs)
	return logs
}

// ProcessTask executes the Propose -> ValidateViaGRPC -> ExecuteOrRollback pipeline.
func (w *ExecutionWorker) ProcessTask(ctx context.Context, task *TaskProposal) error {
	if task == nil {
		return errors.New("task proposal cannot be nil")
	}

	req := &metacoregrpc.ActionEvaluationRequest{
		AgentID:         task.AgentID,
		SessionID:       task.SessionID,
		Goal:            task.Goal,
		PlanSteps:       task.PlanSteps,
		ExpectedOutcome: task.ExpectedOutcome,
		RiskProfile:     task.RiskProfile,
		ResourceCost:    task.ResourceCost,
		Confidence:      task.Confidence,
	}

	// Step 1: ValidateViaGRPC
	// Fail-Closed policy: Any RPC or network failure is treated as execution refusal.
	if w.client == nil {
		return w.handleFailClosed(ctx, task, "Validator client is nil")
	}

	resp, err := w.client.ValidateAction(ctx, req)
	if err != nil {
		return w.handleFailClosed(ctx, task, fmt.Sprintf("RPC network error: %v", err))
	}

	if resp == nil {
		return w.handleFailClosed(ctx, task, "Received nil response from validator")
	}

	// Step 2: Invariant check - execution requires ALLOCATION_DECISION_APPROVED
	isApproved := resp.Decision == metacoregrpc.AllocationDecisionApproved
	var kStatus string
	var reason string

	if resp.KillSwitchOutput != nil {
		kStatus = resp.KillSwitchOutput.Status
		reason = resp.KillSwitchOutput.Reason
	} else {
		reason = fmt.Sprintf("Decision: %s", resp.Decision)
	}

	if !isApproved || kStatus == "rejected" || kStatus == "emergency_stop" {
		// Physical execution blocked! Execute RollbackHandler
		rollbackErr := w.executeRollback(ctx, task, reason)

		w.recordAudit(AuditLogEntry{
			Timestamp:  time.Now(),
			TaskID:     task.ID,
			AgentID:    task.AgentID,
			Decision:   resp.Decision,
			Approved:   false,
			Executed:   false,
			RolledBack: true,
			Reason:     fmt.Sprintf("Execution blocked by Kill-Switch (%s): %s", resp.Decision, reason),
			Error:      fmt.Sprintf("%v", rollbackErr),
		})

		return fmt.Errorf("%w: %s", ErrExecutionRejected, reason)
	}

	// Step 3: Approved - Execute action
	var execErr error
	if task.ExecuteFn != nil {
		execErr = task.ExecuteFn(ctx)
	}

	if execErr != nil {
		// Execution failed at actuator level -> Rollback
		_ = w.executeRollback(ctx, task, fmt.Sprintf("Actuator execution failed: %v", execErr))

		w.recordAudit(AuditLogEntry{
			Timestamp:  time.Now(),
			TaskID:     task.ID,
			AgentID:    task.AgentID,
			Decision:   resp.Decision,
			Approved:   true,
			Executed:   false,
			RolledBack: true,
			Reason:     "Actuator execution failed, rollback executed",
			Error:      execErr.Error(),
		})
		return execErr
	}

	// Success
	w.recordAudit(AuditLogEntry{
		Timestamp:  time.Now(),
		TaskID:     task.ID,
		AgentID:    task.AgentID,
		Decision:   resp.Decision,
		Approved:   true,
		Executed:   true,
		RolledBack: false,
		Reason:     "Action approved and successfully executed",
	})

	return nil
}

func (w *ExecutionWorker) handleFailClosed(ctx context.Context, task *TaskProposal, reason string) error {
	log.Printf("[META-CORE Worker %s] FAIL-CLOSED ENFORCED for task %s: %s", w.workerID, task.ID, reason)

	_ = w.executeRollback(ctx, task, fmt.Sprintf("Fail-Closed policy enforced: %s", reason))

	w.recordAudit(AuditLogEntry{
		Timestamp:  time.Now(),
		TaskID:     task.ID,
		AgentID:    task.AgentID,
		Decision:   metacoregrpc.AllocationDecisionRejected,
		Approved:   false,
		Executed:   false,
		RolledBack: true,
		Reason:     fmt.Sprintf("Fail-Closed Policy: %s", reason),
		Error:      ErrFailClosed.Error(),
	})

	return fmt.Errorf("%w: %s", ErrFailClosed, reason)
}

func (w *ExecutionWorker) executeRollback(ctx context.Context, task *TaskProposal, reason string) error {
	log.Printf("[META-CORE Worker %s] INITIATING ROLLBACK for task %s (Reason: %s)", w.workerID, task.ID, reason)
	if task.RollbackFn != nil {
		return task.RollbackFn(ctx)
	}
	return nil
}

func (w *ExecutionWorker) recordAudit(entry AuditLogEntry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.auditLogs = append(w.auditLogs, entry)
}
