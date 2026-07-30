# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26: META-CORE Validation Pipeline Vulnerabilities Fixed

### Discovery & Risk Assessment
We identified critical inputs within the META-CORE three-layer constraint system (`internal/metacore/validator.go`) that lacked sanitization and strict validation checks. Specifically:
1. **Nil Pointer Dereference**: If a nil `DecisionPacket` was passed to `ValidatePipeline`, it would cause a panic / crash of the agent environment due to pointer dereferencing when evaluating `packet.Goal`, `packet.Plan`, etc.
2. **NaN and Numeric Boundary Bypass**: The `Confidence` score (from `DecisionPacket`) and `VulnerabilityScore` (from `EmpathyOutput`) are crucial metrics for triggering system degradation modes (such as *Conservative Planning* and *Empathy Override*). If an attacker or compromised agent sent out-of-bounds metrics (negative numbers, numbers greater than 1.0, or `NaN`), they could successfully bypass these safety triggers or cause unexpected validation behavior.

### Resolution
1. **Nil Checking**: Explicitly added a `packet == nil` validation checkpoint returning `ErrNilPacket` securely.
2. **Numeric Range and float Verification**: Checked `math.IsNaN(...)` as well as `<= 0.0` and `>= 1.0` boundaries for both `packet.Confidence` and `empathy.VulnerabilityScore`, returning `ErrInvalidConfidence` and `ErrInvalidVulnerability` respectively.
3. **Comprehensive Invariants Protection**: Created robust unit tests to ensure all out-of-bounds conditions correctly fail with specific validation errors.
