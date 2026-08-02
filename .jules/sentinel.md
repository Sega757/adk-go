# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-02-15 - NaN and Out-of-Bound Float Validation Bypass in META-CORE R-E-K
**Vulnerability:** The META-CORE constraint validation pipeline evaluated float64 metrics (Confidence and VulnerabilityScore) directly without sanitizing NaN (Not-a-Number) values. Because all relational comparisons (e.g., <, >) against NaN evaluate to false in Go, high-severity or unstable inputs containing NaN bypassed the pipeline's critical degradation and override triggers. Additionally, passing a nil DecisionPacket caused a nil pointer dereference panic, resulting in Denial of Service (DoS).
**Learning:** Floating-point inputs in critical alignment or safety systems must be validated with high rigor. Standard comparison operators do not protect against IEEE 754 NaN values, and omitting explicit NaN checks allows malicious or unstable inputs to silently bypass system restrictions. Furthermore, defensive nil pointer checks at function boundaries are vital to prevent panic-driven DoS.
**Prevention:** Strictly validate all float inputs in safety-critical pipelines using math.IsNaN() and assert correct numerical boundaries (e.g., [0.0, 1.0]) before any logic evaluations occur. Guard functions against nil pointer inputs with early-return error patterns.
