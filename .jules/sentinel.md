# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 1. IEEE-754 NaN Bypass Vulnerabilities in Safety Pipelines

### Vulnerability Context
In multi-layered security pipelines such as META-CORE, validation components rely on numeric thresholds (e.g., `Confidence` score, `VulnerabilityScore`) to make deterministic safety decisions or trigger safety overrides/degradation modes.

### The Mechanism of a NaN Bypass
Under the IEEE-754 floating-point standard, any comparison involving `NaN` (Not-a-Number) evaluates to `false`. For example:
- `math.NaN() > 0.7` is `false`.
- `math.NaN() < 0.4` is `false`.

Consequently, if an attacker can inject `NaN` as a confidence score or vulnerability score, any logic of the form:
```go
if vulnerability_score > 0.7 {
    // Force modify or block
}
```
will be completely bypassed because the expression evaluates to `false`. The safety validator will proceed as if the vulnerability score is perfectly safe (below the threshold), rendering the entire guardrail system ineffective.

### Prevention & Remediation Guidelines
1. **Explicit `NaN` Checks**: Before using any floating-point number in security/boundary calculations, always validate it using `math.IsNaN()`.
2. **Strict Range Invariants**: Validate that all percentage-based or normalized metrics (such as confidence or vulnerability scores) reside strictly within their expected bounds (e.g., `[0.0, 1.0]`).
3. **Fail Securely on Invalid Input**: If a floating-point value is `NaN` or out of bounds, reject the entire transaction immediately rather than falling back to default/unvalidated execution states.
4. **Input Nil Checks**: Always ensure pointers to decision packets/structures are validated against `nil` before accessing nested fields, preventing denial-of-service (DoS) crashes from nil pointer dereferences.
