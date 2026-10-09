// Copyright 2026 Google LLC
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

package workflow

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCalculateDelay(t *testing.T) {
	cfg := &RetryConfig{
		InitialDelay:  time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      10 * time.Second,
		Jitter:        0.0, // Deterministic for base testing
	}

	tests := []struct {
		failedAttempts int
		want           time.Duration
	}{
		{failedAttempts: 0, want: 0},
		{failedAttempts: 1, want: time.Second},
		{failedAttempts: 2, want: 2 * time.Second},
		{failedAttempts: 3, want: 4 * time.Second},
		{failedAttempts: 4, want: 8 * time.Second},
		{failedAttempts: 5, want: 10 * time.Second}, // Capped at MaxDelay
		{failedAttempts: 6, want: 10 * time.Second},
	}

	for _, tt := range tests {
		got := CalculateDelay(cfg, tt.failedAttempts)
		if got != tt.want {
			t.Errorf("CalculateDelay(..., %d) = %v, want %v", tt.failedAttempts, got, tt.want)
		}
	}
}

// TestCalculateDelayUnsetBackoffFactor guards that an unset BackoffFactor
// gives a constant delay instead of collapsing to 0.
func TestCalculateDelayUnsetBackoffFactor(t *testing.T) {
	cfg := &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: time.Second,
		// BackoffFactor intentionally unset (zero).
	}

	for attempt := 1; attempt <= 3; attempt++ {
		got := CalculateDelay(cfg, attempt)
		if got != time.Second {
			t.Errorf("CalculateDelay(unset backoff, %d) = %v, want %v", attempt, got, time.Second)
		}
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	cfg := &RetryConfig{
		InitialDelay:  time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      10 * time.Second,
		Jitter:        0.5,
	}

	// Run multiple iterations to verify jitter values consistently stay within bounds [0.5s, 1.5s].
	minExpected := 500 * time.Millisecond
	maxExpected := 1500 * time.Millisecond

	seenMap := make(map[time.Duration]bool)

	for i := 0; i < 100; i++ {
		got := CalculateDelay(cfg, 1)
		if got < minExpected || got > maxExpected {
			t.Errorf("CalculateDelay with jitter iteration %d returned %v, expected in range [%v, %v]", i, got, minExpected, maxExpected)
		}
		if got < 0 {
			t.Errorf("CalculateDelay with jitter returned negative delay: %v", got)
		}
		seenMap[got] = true
	}

	// Non-zero variance check: Ensure crypto/rand produces varying values.
	if len(seenMap) <= 1 {
		t.Errorf("CalculateDelay with jitter produced non-varying results across 100 iterations: %v", seenMap)
	}
}

func TestCalculateDelayWithExtremeJitter(t *testing.T) {
	cfg := &RetryConfig{
		InitialDelay: time.Second,
		Jitter:       2.0, // High jitter multiplier that could yield negative offsets
	}

	for i := 0; i < 100; i++ {
		got := CalculateDelay(cfg, 1)
		if got < 0 {
			t.Fatalf("CalculateDelay returned negative delay %v with extreme jitter", got)
		}
	}
}

func TestCalculateDelayConcurrent(t *testing.T) {
	cfg := &RetryConfig{
		InitialDelay:  time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      10 * time.Second,
		Jitter:        0.2,
	}

	var wg sync.WaitGroup
	const goroutines = 20
	const iterationsPerGoroutine = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterationsPerGoroutine; i++ {
				delay := CalculateDelay(cfg, 2)
				if delay < 0 {
					t.Errorf("CalculateDelay returned negative delay: %v", delay)
				}
			}
		}()
	}
	wg.Wait()
}

func TestSecureFloat64(t *testing.T) {
	for i := 0; i < 100; i++ {
		v, err := secureFloat64()
		if err != nil {
			t.Fatalf("secureFloat64 returned error: %v", err)
		}
		if v < 0.0 || v >= 1.0 {
			t.Fatalf("secureFloat64 returned out-of-range value %f, expected [0.0, 1.0)", v)
		}
	}
}

func TestShouldRetry(t *testing.T) {
	errTest := errors.New("test error")

	tests := []struct {
		name           string
		cfg            *RetryConfig
		err            error
		failedAttempts int
		want           bool
	}{
		{
			name:           "Nil config",
			cfg:            nil,
			err:            errTest,
			failedAttempts: 1,
			want:           false,
		},
		{
			name: "Under max attempts",
			cfg: &RetryConfig{
				MaxAttempts: 3,
				ShouldRetry: func(e error) bool { return true },
			},
			err:            errTest,
			failedAttempts: 1,
			want:           true,
		},
		{
			name:           "Default to true when ShouldRetry is nil",
			cfg:            &RetryConfig{MaxAttempts: 3},
			err:            errTest,
			failedAttempts: 1,
			want:           true,
		},
		{
			name:           "At max attempts",
			cfg:            &RetryConfig{MaxAttempts: 3},
			err:            errTest,
			failedAttempts: 3,
			want:           false,
		},
		{
			name:           "Above max attempts",
			cfg:            &RetryConfig{MaxAttempts: 3},
			err:            errTest,
			failedAttempts: 4,
			want:           false,
		},
		{
			name:           "Zero max attempts (no retry)",
			cfg:            &RetryConfig{MaxAttempts: 0},
			err:            errTest,
			failedAttempts: 1,
			want:           false,
		},
		{
			name: "Predicate allows",
			cfg: &RetryConfig{
				MaxAttempts: 3,
				ShouldRetry: func(e error) bool { return true },
			},
			err:            errTest,
			failedAttempts: 1,
			want:           true,
		},
		{
			name: "Predicate denies",
			cfg: &RetryConfig{
				MaxAttempts: 3,
				ShouldRetry: func(e error) bool { return false },
			},
			err:            errTest,
			failedAttempts: 1,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldRetry(tt.cfg, tt.err, tt.failedAttempts)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}
