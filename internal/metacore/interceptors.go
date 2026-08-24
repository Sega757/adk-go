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
	"log"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
)

// TelemetryMetrics collects Prometheus-ready counts and alignment scores.
type TelemetryMetrics struct {
	mu                      sync.Mutex
	TotalRequests           uint64
	SmartButDangerousCount  uint64
	SafeButUselessCount     uint64
	UnstableArchitectureCnt uint64
	NominalCount            uint64
	LastAlignmentScore      float64
	PathologicalTransitions []string
}

// GlobalMetrics provides a thread-safe registry for telemetry metrics.
var GlobalMetrics = &TelemetryMetrics{
	PathologicalTransitions: make([]string, 0),
}

// ResetMetrics clears accumulated telemetry state for testing.
func (m *TelemetryMetrics) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.StoreUint64(&m.TotalRequests, 0)
	atomic.StoreUint64(&m.SmartButDangerousCount, 0)
	atomic.StoreUint64(&m.SafeButUselessCount, 0)
	atomic.StoreUint64(&m.UnstableArchitectureCnt, 0)
	atomic.StoreUint64(&m.NominalCount, 0)
	m.LastAlignmentScore = 0.0
	m.PathologicalTransitions = make([]string, 0)
}

// RecordEvaluation updates Prometheus metrics counters and logs pathological state transitions.
func (m *TelemetryMetrics) RecordEvaluation(score float64, diag string) {
	atomic.AddUint64(&m.TotalRequests, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.LastAlignmentScore = score

	switch diag {
	case "smart_but_dangerous":
		atomic.AddUint64(&m.SmartButDangerousCount, 1)
		msg := "PATHOLOGICAL_TRANSITION [smart_but_dangerous]: High reasoning competence but low humanitarian safety"
		m.PathologicalTransitions = append(m.PathologicalTransitions, msg)
		log.Printf("[META-CORE Telemetry] WARNING: %s | Score: %.2f", msg, score)
	case "safe_but_useless":
		atomic.AddUint64(&m.SafeButUselessCount, 1)
		msg := "PATHOLOGICAL_TRANSITION [safe_but_useless]: High safety filtering but degraded utility"
		m.PathologicalTransitions = append(m.PathologicalTransitions, msg)
		log.Printf("[META-CORE Telemetry] WARNING: %s | Score: %.2f", msg, score)
	case "unstable_architecture":
		atomic.AddUint64(&m.UnstableArchitectureCnt, 1)
		msg := "PATHOLOGICAL_TRANSITION [unstable_architecture]: Frequent Kill-Switch activations detected"
		m.PathologicalTransitions = append(m.PathologicalTransitions, msg)
		log.Printf("[META-CORE Telemetry] CRITICAL: %s | Score: %.2f", msg, score)
	default:
		atomic.AddUint64(&m.NominalCount, 1)
	}
}

// MetricsInterceptor provides a gRPC UnaryServerInterceptor that captures metrics and logs state transitions.
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		_ = time.Since(start)

		if evalResp, ok := resp.(*ActionEvaluationResponse); ok && evalResp != nil {
			GlobalMetrics.RecordEvaluation(evalResp.AlignmentScore, evalResp.DiagnosticStatus)
		}

		return resp, err
	}
}
