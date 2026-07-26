# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## Vulnerability Discoveries & Mitigations

### 1. META-CORE Input Validation Bypasses and Nil Pointer Dereferences

#### Vulnerability Description
During the analysis of the `internal/metacore/validator.go` component, which implements the META-CORE (R–E–K) three-layer constraint system, two distinct security/stability flaws were identified in the `ValidatePipeline` method:
1. **Nil Pointer Dereference (Denial of Service):** The `packet *DecisionPacket` parameter was dereferenced at the beginning of the function without any validation (e.g., `packet.Goal`). If a caller accidentally passed `nil`, this would trigger an unhandled runtime panic, leading to service crashes and potential Denial of Service (DoS).
2. **NaN Bypass Attacks (Bypassing Security Thresholds):** The validation logic relies on comparing numeric scores (`Confidence` and `VulnerabilityScore`) against security thresholds (e.g., `packet.Confidence < 0.4` or `empathy.VulnerabilityScore > 0.7`). In Go, any comparison with `NaN` (Not-a-Number), except `!=`, evaluates to `false`. Therefore, an attacker or a faulty agent state supplying `NaN` values could successfully bypass constraint thresholds and compromise the security invariants of the system.

#### Mitigation Action
- Integrated robust, strict input validation checks at the very entry point of the `ValidatePipeline` function.
- Implemented an explicit nil check on the `packet` argument, returning a specific `ErrNilPacket` error rather than panicking.
- Utilized Go's built-in `math.IsNaN` package function to check both `packet.Confidence` and `empathy.VulnerabilityScore` for `NaN` states.
- Bound-checked `Confidence` and `VulnerabilityScore` to strictly reside within the valid interval `[0.0, 1.0]`. If any check fails, appropriate, secure errors are returned (`ErrInvalidConfidence` and `ErrInvalidVulnerability`).
- Created a comprehensive set of unit tests in `internal/metacore/validator_test.go` to enforce validation invariants against malicious, malformed, or boundary inputs.
