# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - Bypassing Float-Based Invariant Checks via NaN and Nil Values in META-CORE Pipeline
**Vulnerability:** The META-CORE reasoning-empathy-kill-switch pipeline evaluated floating-point parameters (`Confidence` and `VulnerabilityScore`) for critical safety overrides and modes. However, it lacked input boundary validation and fell victim to NaN (Not-a-Number) comparison bypasses, as well as a potential Denial-of-Service via nil pointer dereference of the `DecisionPacket`.
**Learning:** In Go, any comparison with a `NaN` floating-point value evaluates to `false` (e.g., `NaN > 0.85` or `NaN < 0.4` is false). A malicious actor or compromised agent could craft packets with `NaN` confidence/vulnerability values to completely bypass conservative planning or forced empathy modifications, while keeping the pipeline in nominal execution mode.
**Prevention:** Always validate floating-point inputs against explicit valid ranges (e.g. `!(val >= 0.0 && val <= 1.0)`) before they are processed by conditional logic, and robustly check all input pointers for `nil` before any dereferencing.
