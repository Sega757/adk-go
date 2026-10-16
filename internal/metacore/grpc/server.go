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

package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/adk/v2/internal/metacore"
)

// AllocationDecision constants matching metacore.rek.v1.
type AllocationDecision int32

const (
	AllocationDecisionUnspecified   AllocationDecision = 0
	AllocationDecisionApproved      AllocationDecision = 1
	AllocationDecisionModified      AllocationDecision = 2
	AllocationDecisionRejected      AllocationDecision = 3
	AllocationDecisionEmergencyStop AllocationDecision = 4
)

func (a AllocationDecision) String() string {
	switch a {
	case AllocationDecisionApproved:
		return "ALLOCATION_DECISION_APPROVED"
	case AllocationDecisionModified:
		return "ALLOCATION_DECISION_MODIFIED"
	case AllocationDecisionRejected:
		return "ALLOCATION_DECISION_REJECTED"
	case AllocationDecisionEmergencyStop:
		return "ALLOCATION_DECISION_EMERGENCY_STOP"
	default:
		return "ALLOCATION_DECISION_UNSPECIFIED"
	}
}

// RiskProfile represents ethical, legal, operational, and human impact risks.
type RiskProfile struct {
	Ethical     string `json:"ethical"`
	Legal       string `json:"legal"`
	Operational string `json:"operational"`
	HumanImpact string `json:"human_impact"`
}

// ResourceCost defines resource parameters.
type ResourceCost struct {
	ComputeMillicores uint64  `json:"compute_millicores"`
	BudgetUSD         float64 `json:"budget_usd"`
	TimeoutMS         uint32  `json:"timeout_ms"`
}

// ActionEvaluationRequest is the unary request message.
type ActionEvaluationRequest struct {
	AgentID         string                  `json:"agent_id"`
	SessionID       string                  `json:"session_id"`
	Goal            string                  `json:"goal"`
	PlanSteps       []string                `json:"plan_steps"`
	ExpectedOutcome string                  `json:"expected_outcome"`
	RiskProfile     RiskProfile             `json:"risk_profile"`
	ResourceCost    ResourceCost            `json:"resource_cost"`
	Confidence      float32                 `json:"confidence"`
	EmpathyInput    *metacore.EmpathyOutput `json:"empathy_input,omitempty"`
}

// EmpathyDetails holds Empathy Layer output.
type EmpathyDetails struct {
	Status               string  `json:"status"`
	VulnerabilityScore   float32 `json:"vulnerability_score"`
	ToneModification     string  `json:"tone_modification"`
	IntensityModification string `json:"intensity_modification"`
	DelayMS              uint32  `json:"delay_ms"`
	Reason               string  `json:"reason"`
}

// KillSwitchDetails holds Kill-Switch Layer output.
type KillSwitchDetails struct {
	Status            string   `json:"status"`
	ActivatedTriggers []string `json:"activated_triggers"`
	RollbackRequired  bool     `json:"rollback_required"`
	Reason            string   `json:"reason"`
}

// ActionEvaluationResponse is the unary evaluation response.
type ActionEvaluationResponse struct {
	Decision           AllocationDecision `json:"decision"`
	EmpathyLayerOutput *EmpathyDetails    `json:"empathy_layer_output"`
	KillSwitchOutput   *KillSwitchDetails `json:"kill_switch_output"`
}

// ReasoningStateHeartbeat represents streaming CoT audit heartbeat.
type ReasoningStateHeartbeat struct {
	AgentID            string `json:"agent_id"`
	SessionID          string `json:"session_id"`
	ExecutionStepIndex uint32 `json:"execution_step_index"`
	LogicTrace         string `json:"logic_trace"`
	CurrentMemoryBytes uint64 `json:"current_memory_bytes"`
}

// SecurityDirective is sent back in streaming CoT audit.
type SecurityDirective struct {
	AbortExecution          bool   `json:"abort_execution"`
	ForceRollback           bool   `json:"force_rollback"`
	ModificationInstruction string `json:"modification_instruction"`
	Reason                  string `json:"reason"`
}

// AuditStream defines interface for streaming CoT reasoning audit.
type AuditStream interface {
	grpc.ServerStream
	Recv() (*ReasoningStateHeartbeat, error)
	Send(*SecurityDirective) error
}

// ServerConfig configures MetaCoreValidatorServer.
type ServerConfig struct {
	DefaultTimeout      time.Duration
	MaxVulnerability    float64
	ResourceLimit       float64
	MaxAllowedKTriggers int
	MemoryLimitBytes    uint64
}

// MetaCoreValidatorServer implements the gRPC validation service.
type MetaCoreValidatorServer struct {
	mu        sync.RWMutex
	validator *metacore.Validator
	config    ServerConfig
}

// NewServer initializes a new MetaCoreValidatorServer.
func NewServer(cfg ServerConfig) *MetaCoreValidatorServer {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 10 * time.Millisecond
	}
	if cfg.MaxVulnerability == 0 {
		cfg.MaxVulnerability = 0.7
	}
	if cfg.ResourceLimit == 0 {
		cfg.ResourceLimit = 100.0
	}
	if cfg.MaxAllowedKTriggers == 0 {
		cfg.MaxAllowedKTriggers = 5
	}
	if cfg.MemoryLimitBytes == 0 {
		cfg.MemoryLimitBytes = 1024 * 1024 * 1024 // 1GB default
	}

	v := metacore.NewValidator(cfg.MaxVulnerability, cfg.ResourceLimit, cfg.MaxAllowedKTriggers)

	return &MetaCoreValidatorServer{
		validator: v,
		config:    cfg,
	}
}

// ValidateAction handles unary action validation under strict context timeouts.
func (s *MetaCoreValidatorServer) ValidateAction(ctx context.Context, req *ActionEvaluationRequest) (*ActionEvaluationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, s.config.DefaultTimeout)
		defer cancel()
	}

	type evalResult struct {
		resp *ActionEvaluationResponse
		err  error
	}

	ch := make(chan evalResult, 1)

	go func() {
		s.mu.RLock()
		mode := s.validator.CurrentMode
		s.mu.RUnlock()

		if mode == metacore.ModeSafeMode {
			ch <- evalResult{
				resp: &ActionEvaluationResponse{
					Decision: AllocationDecisionRejected,
					KillSwitchOutput: &KillSwitchDetails{
						Status:            "rejected",
						ActivatedTriggers: []string{"resource_limit"},
						RollbackRequired:  true,
						Reason:            "System locked in Safe Mode due to repeated K triggers",
					},
				},
				err: status.Error(codes.Unavailable, "system operating under Safe Mode lockout"),
			}
			return
		}

		packet := &metacore.DecisionPacket{
			Goal:            req.Goal,
			Plan:            req.PlanSteps,
			ExpectedOutcome: req.ExpectedOutcome,
			RiskProfile: metacore.RiskProfile{
				Ethical:     req.RiskProfile.Ethical,
				Legal:       req.RiskProfile.Legal,
				Operational: req.RiskProfile.Operational,
				HumanImpact: req.RiskProfile.HumanImpact,
			},
			ResourceCost: metacore.ResourceCost{
				Compute: fmt.Sprintf("%d millicores", req.ResourceCost.ComputeMillicores),
				Budget:  fmt.Sprintf("%.2f USD", req.ResourceCost.BudgetUSD),
				Time:    fmt.Sprintf("%d ms", req.ResourceCost.TimeoutMS),
			},
			Confidence: float64(req.Confidence),
		}

		empathyInput := req.EmpathyInput
		if empathyInput == nil {
			empathyInput = &metacore.EmpathyOutput{
				Status:             "pass",
				VulnerabilityScore: 0.1,
				Reason:             "Default gRPC evaluation",
			}
		}

		s.mu.Lock()
		kOut, err := s.validator.ValidatePipeline(packet, empathyInput)
		s.mu.Unlock()

		decision := AllocationDecisionApproved
		var kDetails *KillSwitchDetails
		var eDetails *EmpathyDetails

		if empathyInput != nil {
			eDetails = &EmpathyDetails{
				Status:                empathyInput.Status,
				VulnerabilityScore:    float32(empathyInput.VulnerabilityScore),
				ToneModification:      empathyInput.Modifications.Tone,
				IntensityModification: empathyInput.Modifications.Intensity,
				Reason:                empathyInput.Reason,
			}
		}

		if kOut != nil {
			kDetails = &KillSwitchDetails{
				Status:            kOut.Status,
				ActivatedTriggers: kOut.Trigger,
				RollbackRequired:  kOut.Rollback,
				Reason:            kOut.Reason,
			}

			switch kOut.Status {
			case "rejected":
				decision = AllocationDecisionRejected
			case "emergency_stop":
				decision = AllocationDecisionEmergencyStop
			default:
				if empathyInput != nil && empathyInput.Status == "modify" {
					decision = AllocationDecisionModified
				}
			}
		}

		resp := &ActionEvaluationResponse{
			Decision:           decision,
			EmpathyLayerOutput: eDetails,
			KillSwitchOutput:   kDetails,
		}

		if errors.Is(err, metacore.ErrKAbsoluteBlock) {
			ch <- evalResult{
				resp: resp,
				err:  status.Error(codes.PermissionDenied, fmt.Sprintf("K-Switch Block: %s", kOut.Reason)),
			}
			return
		}

		if err != nil {
			ch <- evalResult{
				resp: nil,
				err:  status.Error(codes.InvalidArgument, err.Error()),
			}
			return
		}

		ch <- evalResult{resp: resp, err: nil}
	}()

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "action evaluation timed out")
	case res := <-ch:
		return res.resp, res.err
	}
}

// AuditReasoningStream performs continuous real-time CoT stream auditing with anomaly early-stop.
func (s *MetaCoreValidatorServer) AuditReasoningStream(stream AuditStream) error {
	var lastStepTime time.Time

	for {
		hb, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		now := time.Now()
		if !lastStepTime.IsZero() {
			latency := now.Sub(lastStepTime)
			// Log warning if CoT step latency exceeds 100ms
			if latency > 100*time.Millisecond {
				log.Printf("[META-CORE CoT Audit] WARNING: High CoT step latency: %v for step %d", latency, hb.ExecutionStepIndex)
			}
		}
		lastStepTime = now

		// Real-time anomaly detection in CoT reasoning trace
		traceLower := strings.ToLower(hb.LogicTrace)
		hasAnomaly := false
		var anomalyReason string

		if strings.Contains(traceLower, "bypass") || strings.Contains(traceLower, "disable safety") {
			hasAnomaly = true
			anomalyReason = "Goal drift or safety bypass attempt detected in reasoning trace"
		} else if strings.Contains(traceLower, "harm") || strings.Contains(traceLower, "exploit") || strings.Contains(traceLower, "cheat logs") {
			hasAnomaly = true
			anomalyReason = "Ethical boundary or reward hacking anomaly in CoT trace"
		} else if hb.CurrentMemoryBytes > s.config.MemoryLimitBytes {
			hasAnomaly = true
			anomalyReason = fmt.Sprintf("Abnormal memory usage detected in CoT trace: %d bytes", hb.CurrentMemoryBytes)
		}

		if hasAnomaly {
			directive := &SecurityDirective{
				AbortExecution:          true,
				ForceRollback:           true,
				ModificationInstruction: "HALT_IMMEDIATE",
				Reason:                  anomalyReason,
			}
			if sendErr := stream.Send(directive); sendErr != nil {
				return sendErr
			}
			return status.Error(codes.Aborted, fmt.Sprintf("CoT Audit Early Stop: %s", anomalyReason))
		}

		// Normal heartbeat ack
		directive := &SecurityDirective{
			AbortExecution:          false,
			ForceRollback:           false,
			ModificationInstruction: "NONE",
			Reason:                  "CoT step verified",
		}
		if sendErr := stream.Send(directive); sendErr != nil {
			return sendErr
		}
	}
}

// UnaryPanicRecoveryInterceptor provides gRPC panic recovery with error sanitization for unary calls.
func UnaryPanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[META-CORE gRPC] RECOVERED PANIC in %s: %v\nStack:\n%s", info.FullMethod, r, string(debug.Stack()))
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// StreamPanicRecoveryInterceptor provides gRPC panic recovery with error sanitization for streaming calls.
func StreamPanicRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[META-CORE gRPC] RECOVERED PANIC in %s: %v\nStack:\n%s", info.FullMethod, r, string(debug.Stack()))
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}

// StartServer launches a gRPC listener with panic interceptors.
func StartServer(lis net.Listener, cfg ServerConfig) (*grpc.Server, *MetaCoreValidatorServer, error) {
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(UnaryPanicRecoveryInterceptor(), metacore.MetricsInterceptor()),
		grpc.ChainStreamInterceptor(StreamPanicRecoveryInterceptor()),
	)

	srv := NewServer(cfg)

	go func() {
		_ = grpcSrv.Serve(lis)
	}()

	return grpcSrv, srv, nil
}
