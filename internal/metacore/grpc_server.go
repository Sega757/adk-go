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

package metacore

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServerConfig defines production parameters for the rekv1 gRPC service.
type GRPCServerConfig struct {
	DefaultTimeout      time.Duration
	MaxVulnerability    float64
	ResourceLimit       float64
	MaxAllowedKTriggers int
}

// Server implements the rekv1 META-CORE gRPC validation service.
type Server struct {
	mu        sync.RWMutex
	validator *Validator
	config    GRPCServerConfig
}

// NewServer initializes a new META-CORE gRPC server with the given configuration.
func NewServer(cfg GRPCServerConfig) *Server {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 50 * time.Millisecond
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

	v := NewValidator(cfg.MaxVulnerability, cfg.ResourceLimit, cfg.MaxAllowedKTriggers)

	return &Server{
		validator: v,
		config:    cfg,
	}
}

// ActionEvaluationRequest mirrors the protobuf ActionEvaluationRequest struct.
type ActionEvaluationRequest struct {
	AgentID         string                 `json:"agent_id"`
	SessionID       string                 `json:"session_id"`
	Goal            string                 `json:"goal"`
	PlanSteps       []string               `json:"plan_steps"`
	ExpectedOutcome string                 `json:"expected_outcome"`
	RiskProfile     RiskProfile            `json:"risk_profile"`
	ResourceCost    ResourceCost           `json:"resource_cost"`
	Confidence      float64                `json:"confidence"`
	EmpathyInput    *EmpathyOutput         `json:"empathy_input,omitempty"`
}

// ActionEvaluationResponse mirrors the protobuf ActionEvaluationResponse struct.
type ActionEvaluationResponse struct {
	Decision           string            `json:"decision"` // APPROVED, MODIFIED, REJECTED, EMERGENCY_STOP
	EmpathyLayerOutput *EmpathyOutput    `json:"empathy_layer_output"`
	KillSwitchOutput   *KillSwitchOutput `json:"kill_switch_output"`
	AlignmentScore     float64           `json:"alignment_score"`
	DiagnosticStatus   string            `json:"diagnostic_status"`
	CurrentMode        DegradationMode   `json:"current_mode"`
}

// ValidateAction processes an action evaluation request under context deadlines and timeouts.
func (s *Server) ValidateAction(ctx context.Context, req *ActionEvaluationRequest) (*ActionEvaluationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Apply timeout from context or default
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, s.config.DefaultTimeout)
		defer cancel()
	}

	// Create channel for asynchronous evaluation to respect deadline
	type evalResult struct {
		resp *ActionEvaluationResponse
		err  error
	}

	ch := make(chan evalResult, 1)

	go func() {
		s.mu.RLock()
		mode := s.validator.CurrentMode
		s.mu.RUnlock()

		// If validator is in Safe Mode, immediately reject dynamic tool executions
		if mode == ModeSafeMode {
			ch <- evalResult{
				resp: &ActionEvaluationResponse{
					Decision:         "REJECTED",
					AlignmentScore:   0.0,
					DiagnosticStatus: "safe_mode_lockout",
					CurrentMode:      ModeSafeMode,
					KillSwitchOutput: &KillSwitchOutput{
						Status:   "rejected",
						Reason:   "System locked in Safe Mode due to repeated K triggers",
						Trigger:  []string{"resource_limit"},
						Rollback: true,
					},
				},
				err: status.Error(codes.Unavailable, "system is operating under Safe Mode lockout"),
			}
			return
		}

		packet := &DecisionPacket{
			Goal:            req.Goal,
			Plan:            req.PlanSteps,
			ExpectedOutcome: req.ExpectedOutcome,
			RiskProfile:     req.RiskProfile,
			ResourceCost:    req.ResourceCost,
			Confidence:      req.Confidence,
		}

		empathyInput := req.EmpathyInput
		if empathyInput == nil {
			empathyInput = &EmpathyOutput{
				Status:             "pass",
				VulnerabilityScore: 0.1,
				Reason:             "Default gRPC evaluation",
			}
		}

		// Calculate initial alignment score
		score, diag := s.validator.EvaluateAlignment(req.Confidence, 1.0-empathyInput.VulnerabilityScore, 1.0)

		s.mu.Lock()
		kOut, err := s.validator.ValidatePipeline(packet, empathyInput)
		currentMode := s.validator.CurrentMode
		s.mu.Unlock()

		decision := "APPROVED"
		if kOut != nil {
			switch kOut.Status {
			case "rejected":
				decision = "REJECTED"
			case "emergency_stop":
				decision = "EMERGENCY_STOP"
			default:
				if empathyInput.Status == "modify" {
					decision = "MODIFIED"
				}
			}
		}

		resp := &ActionEvaluationResponse{
			Decision:           decision,
			EmpathyLayerOutput: empathyInput,
			KillSwitchOutput:   kOut,
			AlignmentScore:     score,
			DiagnosticStatus:   diag,
			CurrentMode:        currentMode,
		}

		if errors.Is(err, ErrKAbsoluteBlock) {
			ch <- evalResult{resp: resp, err: status.Error(codes.PermissionDenied, fmt.Sprintf("K-Switch Block: %s", kOut.Reason))}
			return
		}
		if err != nil {
			ch <- evalResult{resp: nil, err: status.Error(codes.InvalidArgument, err.Error())}
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

// StartGRPCServer launches a production-ready gRPC listener.
func StartGRPCServer(lis net.Listener, cfg GRPCServerConfig, interceptors ...grpc.UnaryServerInterceptor) (*grpc.Server, *Server, error) {
	opts := []grpc.ServerOption{}
	if len(interceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))
	}

	grpcSrv := grpc.NewServer(opts...)
	srv := NewServer(cfg)

	go func() {
		_ = grpcSrv.Serve(lis)
	}()

	return grpcSrv, srv, nil
}
