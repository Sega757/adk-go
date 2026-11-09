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

package grpc_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	metacoregrpc "google.golang.org/adk/v2/internal/metacore/grpc"
)

type mockStream struct {
	grpc.ServerStream
	recvCh chan *metacoregrpc.ReasoningStateHeartbeat
	sendCh chan *metacoregrpc.SecurityDirective
	ctx    context.Context
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) Recv() (*metacoregrpc.ReasoningStateHeartbeat, error) {
	hb, ok := <-m.recvCh
	if !ok {
		return nil, fmt.Errorf("stream closed")
	}
	return hb, nil
}

func (m *mockStream) Send(sd *metacoregrpc.SecurityDirective) error {
	m.sendCh <- sd
	return nil
}

func TestGRPCServer_ValidateAction_Integration(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()

	grpcSrv, srv, err := metacoregrpc.StartServer(lis, metacoregrpc.ServerConfig{
		DefaultTimeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer grpcSrv.Stop()

	req := &metacoregrpc.ActionEvaluationRequest{
		AgentID:         "agent-01",
		SessionID:       "sess-01",
		Goal:            "Analyze user request safely",
		PlanSteps:       []string{"read config", "summarize"},
		ExpectedOutcome: "valid summary",
		Confidence:      0.9,
	}

	resp, err := srv.ValidateAction(context.Background(), req)
	if err != nil {
		t.Fatalf("ValidateAction failed: %v", err)
	}

	if resp.Decision != metacoregrpc.AllocationDecisionApproved {
		t.Errorf("Expected decision APPROVED, got %v", resp.Decision)
	}
}

func TestGRPCServer_AuditReasoningStream_EarlyStop(t *testing.T) {
	srv := metacoregrpc.NewServer(metacoregrpc.ServerConfig{
		MemoryLimitBytes: 1024 * 1024,
	})

	recvCh := make(chan *metacoregrpc.ReasoningStateHeartbeat, 5)
	sendCh := make(chan *metacoregrpc.SecurityDirective, 5)

	stream := &mockStream{
		recvCh: recvCh,
		sendCh: sendCh,
		ctx:    context.Background(),
	}

	// Normal heartbeat
	recvCh <- &metacoregrpc.ReasoningStateHeartbeat{
		AgentID:            "agent-01",
		SessionID:          "sess-01",
		ExecutionStepIndex: 1,
		LogicTrace:         "Step 1: Normal planning step",
		CurrentMemoryBytes: 1024,
	}

	// Anomaly heartbeat: safety bypass attempt
	recvCh <- &metacoregrpc.ReasoningStateHeartbeat{
		AgentID:            "agent-01",
		SessionID:          "sess-01",
		ExecutionStepIndex: 2,
		LogicTrace:         "Step 2: Attempt to bypass safety checks",
		CurrentMemoryBytes: 2048,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.AuditReasoningStream(stream)
	}()

	// First Directive: Approved
	dir1 := <-sendCh
	if dir1.AbortExecution {
		t.Errorf("Expected dir1 AbortExecution=false, got true")
	}

	// Second Directive: Aborted
	dir2 := <-sendCh
	if !dir2.AbortExecution || !dir2.ForceRollback {
		t.Errorf("Expected dir2 AbortExecution=true and ForceRollback=true, got %v", dir2)
	}

	streamErr := <-errCh
	if status.Code(streamErr) != codes.Aborted {
		t.Errorf("Expected stream error code Aborted, got %v", streamErr)
	}
}

func TestGRPCServer_PanicRecovery(t *testing.T) {
	interceptor := metacoregrpc.UnaryPanicRecoveryInterceptor()

	handler := func(ctx context.Context, req any) (any, error) {
		panic("simulated server panic")
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/metacore.rek.v1.MetaCoreValidator/ValidateAction"}
	_, err := interceptor(context.Background(), nil, info, handler)

	if err == nil {
		t.Fatalf("Expected error from panic recovery interceptor, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal || st.Message() != "internal server error" {
		t.Errorf("Expected sanitized internal server error, got %v", err)
	}
}

func BenchmarkValidateAction_EndToEnd(b *testing.B) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()

	grpcSrv, srv, err := metacoregrpc.StartServer(lis, metacoregrpc.ServerConfig{
		DefaultTimeout: 50 * time.Millisecond,
	})
	if err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer grpcSrv.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		b.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	req := &metacoregrpc.ActionEvaluationRequest{
		AgentID:         "bench-agent",
		SessionID:       "bench-sess",
		Goal:            "Perform benchmark operation",
		PlanSteps:       []string{"step 1", "step 2"},
		ExpectedOutcome: "benchmark output",
		Confidence:      0.95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := srv.ValidateAction(context.Background(), req)
		if err != nil {
			b.Fatalf("ValidateAction failed: %v", err)
		}
		if resp.Decision != metacoregrpc.AllocationDecisionApproved {
			b.Fatalf("Expected approved decision, got %v", resp.Decision)
		}
	}
}
