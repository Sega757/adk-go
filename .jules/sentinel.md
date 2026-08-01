# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26: META-CORE Validation Vulnerabilities Resolved

### Vulnerability Discoveries
1. **Nil Pointer Dereference**:
   - In `internal/metacore/validator.go`, the `ValidatePipeline` function evaluated `packet` without checking if it was a `nil` pointer first (e.g., calling `strings.ToLower(packet.Goal)`). This could cause the validation service to crash, leading to a denial of service (DoS).
2. **Missing Input Validation & Float Bypass (`NaN` Injection)**:
   - Both `Confidence` and `VulnerabilityScore` are represented as floating-point numbers (`float64`).
   - The validation engine didn't verify if these fields were valid numbers or within their boundaries `[0.0, 1.0]`. Specifically, an attacker or compromised logical/empathy model could emit `NaN` (Not-a-Number) values, which bypass conventional comparison constraints (e.g. `NaN < 0.4` or `NaN > 0.7` are both false), thereby bypassing degradation triggers and safety invariant controls.

### Mitigation & Prevention Guidelines
1. **Always Validate Pointers**: Ensure that all incoming request/packet arguments are thoroughly checked for `nil` before accessing fields or methods.
2. **Strict Float Range Checking & NaN Handling**: When handling floats that represent probabilities or metric scores (like Confidence or Vulnerability score), validate using `math.IsNaN(val)` and verify they strictly reside in the `[0.0, 1.0]` range.
