# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## Vulnerability Discoveries

### 1. Nil Pointer Dereference in META-CORE Validation Pipeline
- **Package**: `internal/metacore`
- **Component**: `ValidatePipeline` in `validator.go`
- **Description**: The function accepted `packet *DecisionPacket` and accessed nested properties like `packet.Goal` without checking if the pointer was `nil`. This could be exploited to trigger a nil pointer dereference, crashing the agent runtime process or causing service-wide denial of service (DoS).
- **Remediation**: Added an explicit check for `packet == nil` at the beginning of `ValidatePipeline`, returning a new `ErrNilDecisionPacket` error.

### 2. NaN Bypass Attack on Numeric Safety Thresholds
- **Package**: `internal/metacore`
- **Component**: `ValidatePipeline` in `validator.go`
- **Description**: The system relied on floating-point comparison thresholds to enforce safety constraints, e.g., `vulnerability_score > 0.7` to trigger a mandatory modify/block, or `confidence < 0.4` to trigger conservative planning. In IEEE 754 floating-point arithmetic, comparisons with `NaN` (Not-a-Number) always evaluate to `false` (i.e., `NaN > 0.7` is false). An adversary or corrupted agent state could exploit this by injecting `NaN` into these fields, allowing high-risk activities to bypass the safety layers completely without triggering validations or audit trails.
- **Remediation**: Added checks using Go's `math.IsNaN` to reject `NaN` inputs.

### 3. Out-of-Bounds/Inf Metric Spoofing
- **Package**: `internal/metacore`
- **Component**: `ValidatePipeline` in `validator.go`
- **Description**: Scores like `Confidence` and `VulnerabilityScore` are mathematically bound to the range `[0.0, 1.0]`. An adversary could pass extreme out-of-bounds inputs or infinite values (`+Inf`, `-Inf`) to bypass validation logic or distort the mathematical alignment score calculation ($A = rVal * eSaf * kSaf$).
- **Remediation**: Integrated validation checks using `math.IsInf` and range boundary comparisons to restrict both `Confidence` and `VulnerabilityScore` strictly to `[0.0, 1.0]`.

---

## Prevention & Hardening Guidelines
1. **Never Assume Valid Pointer Inputs**: For any public or internal orchestration boundary, explicitly check pointer arguments for `nil` before accessing their properties.
2. **Sanitize Floating-Point Metrics**:
   - For safety-critical numbers, check against `NaN` (`math.IsNaN(val)`) and infinity (`math.IsInf(val, 0)`).
   - Enforce rigorous bounds checks (`val < 0.0 || val > 1.0`) to avoid logical bypasses or overflow conditions.
3. **Fail Securely**: When validation fails, immediately return a specific validation error, halt execution, and reject the proposed plan rather than attempting to auto-correct or proceed with default values.
