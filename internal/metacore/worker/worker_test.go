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

package worker_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	metacoregrpc "google.golang.org/adk/v2/internal/metacore/grpc"
	"google.golang.org/adk/v2/internal/metacore/worker"
)

type mockClient struct {
	validateFn func(ctx context.Context, req *metacoregrpc.ActionEvaluationRequest) (*metacoregrpc.ActionEvaluationResponse, error)
}

func (m *mockClient) ValidateAction(ctx context.Context, req *metacoregrpc.ActionEvaluationRequest) (*metacoregrpc.ActionEvaluationResponse, error) {
	if m.validateFn != nil {
		return m.validateFn(ctx, req)
	}
	return &metacoregrpc.ActionEvaluationResponse{
		Decision: metacoregrpc.AllocationDecisionApproved,
		KillSwitchOutput: &metacoregrpc.KillSwitchDetails{
			Status: "approved",
			Reason: "All invariants satisfied",
		},
	}, nil
}

func TestWorker_Integration_Approved(t *testing.T) {
	client := &mockClient{}
	w := worker.NewExecutionWorker("worker-01", client)

	var executed uint32
	var rolledBack uint32

	task := &worker.TaskProposal{
		ID:              "task-01",
		AgentID:         "agent-01",
		SessionID:       "sess-01",
		Goal:            "Process valid user task",
		PlanSteps:       []string{"step1", "step2"},
		ExpectedOutcome: "processed",
		Confidence:      0.9,
		ExecuteFn: func(ctx context.Context) error {
			atomic.StoreUint32(&executed, 1)
			return nil
		},
		RollbackFn: func(ctx context.Context) error {
			atomic.StoreUint32(&rolledBack, 1)
			return nil
		},
	}

	err := w.ProcessTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Expected successful processing, got error: %v", err)
	}

	if atomic.LoadUint32(&executed) != 1 {
		t.Errorf("Expected task to execute, but executed == 0")
	}
	if atomic.LoadUint32(&rolledBack) != 0 {
		t.Errorf("Expected no rollback, but rolledBack == 1")
	}

	logs := w.AuditLogs()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 audit log entry, got %d", len(logs))
	}
	if !logs[0].Approved || !logs[0].Executed {
		t.Errorf("Expected audit log to reflect approved & executed task, got: %+v", logs[0])
	}
}

func TestWorker_Rejection_ViolatingPacket(t *testing.T) {
	client := &mockClient{
		validateFn: func(ctx context.Context, req *metacoregrpc.ActionEvaluationRequest) (*metacoregrpc.ActionEvaluationResponse, error) {
			return &metacoregrpc.ActionEvaluationResponse{
				Decision: metacoregrpc.AllocationDecisionRejected,
				KillSwitchOutput: &metacoregrpc.KillSwitchDetails{
					Status:            "rejected",
					ActivatedTriggers: []string{"ethical_violation"},
					RollbackRequired:  true,
					Reason:            "6-Field Violation: Physical Safety breached",
				},
			}, nil
		},
	}

	w := worker.NewExecutionWorker("worker-01", client)

	var executed uint32
	var rolledBack uint32

	task := &worker.TaskProposal{
		ID:              "task-violating",
		AgentID:         "agent-01",
		Goal:            "harm user or system",
		PlanSteps:       []string{"kill process"},
		Confidence:      0.9,
		ExecuteFn: func(ctx context.Context) error {
			atomic.StoreUint32(&executed, 1)
			return nil
		},
		RollbackFn: func(ctx context.Context) error {
			atomic.StoreUint32(&rolledBack, 1)
			return nil
		},
	}

	err := w.ProcessTask(context.Background(), task)
	if err == nil {
		t.Fatalf("Expected error for rejected task, got nil")
	}

	if !errors.Is(err, worker.ErrExecutionRejected) {
		t.Errorf("Expected ErrExecutionRejected, got %v", err)
	}

	if atomic.LoadUint32(&executed) != 0 {
		t.Errorf("INVARIANT BREACH: Physical execution was NOT blocked on rejected task!")
	}
	if atomic.LoadUint32(&rolledBack) != 1 {
		t.Errorf("Expected RollbackFn to be called on rejected task, but rolledBack == 0")
	}

	logs := w.AuditLogs()
	if len(logs) != 1 || logs[0].Approved || logs[0].Executed || !logs[0].RolledBack {
		t.Errorf("Expected audit log to record rejected & rolled back task, got: %+v", logs)
	}
}

func TestWorker_NetworkFailure_FailClosed(t *testing.T) {
	client := &mockClient{
		validateFn: func(ctx context.Context, req *metacoregrpc.ActionEvaluationRequest) (*metacoregrpc.ActionEvaluationResponse, error) {
			return nil, fmt.Errorf("connection refused: gRPC server unreachable")
		},
	}

	w := worker.NewExecutionWorker("worker-01", client)

	var executed uint32
	var rolledBack uint32

	task := &worker.TaskProposal{
		ID:         "task-offline",
		AgentID:    "agent-01",
		Goal:       "perform action while server offline",
		Confidence: 0.9,
		ExecuteFn: func(ctx context.Context) error {
			atomic.StoreUint32(&executed, 1)
			return nil
		},
		RollbackFn: func(ctx context.Context) error {
			atomic.StoreUint32(&rolledBack, 1)
			return nil
		},
	}

	err := w.ProcessTask(context.Background(), task)
	if err == nil {
		t.Fatalf("Expected error on network failure, got nil")
	}

	if !errors.Is(err, worker.ErrFailClosed) {
		t.Errorf("Expected ErrFailClosed, got %v", err)
	}

	if atomic.LoadUint32(&executed) != 0 {
		t.Errorf("INVARIANT BREACH: Execution occurred despite RPC network failure!")
	}
	if atomic.LoadUint32(&rolledBack) != 1 {
		t.Errorf("Expected RollbackFn to be called on fail-closed network failure, got 0")
	}

	logs := w.AuditLogs()
	if len(logs) != 1 || logs[0].Approved || logs[0].Executed || !logs[0].RolledBack {
		t.Errorf("Expected audit log to record fail-closed rejection & rollback, got: %+v", logs)
	}
}

func TestWorker_ActuatorFailure_Rollback(t *testing.T) {
	client := &mockClient{}
	w := worker.NewExecutionWorker("worker-01", client)

	var rolledBack uint32

	task := &worker.TaskProposal{
		ID:      "task-actuator-fail",
		AgentID: "agent-01",
		Goal:    "execute task",
		ExecuteFn: func(ctx context.Context) error {
			return fmt.Errorf("actuator hardware error")
		},
		RollbackFn: func(ctx context.Context) error {
			atomic.StoreUint32(&rolledBack, 1)
			return nil
		},
	}

	err := w.ProcessTask(context.Background(), task)
	if err == nil || err.Error() != "actuator hardware error" {
		t.Fatalf("Expected actuator hardware error, got %v", err)
	}

	if atomic.LoadUint32(&rolledBack) != 1 {
		t.Errorf("Expected rollback when actuator fails, got %d", rolledBack)
	}
}
