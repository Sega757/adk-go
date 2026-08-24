# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.

## 2026-07-27 - [Strict META-CORE Degradation Mode Enforcement]
**Vulnerability:** Although degradation modes (Safe Mode, Empathy Override, Conservative Planning) were defined and transitioned to under certain conditions, their actual execution constraints were never enforced. This allowed reasoning engines to bypass safety limits (e.g., executing arbitrary dynamic tools or exceeding decision depths) even when the system was supposed to be degraded to safe or conservative modes.
**Learning:** Security state machines must enforce their corresponding behavior transitions actively in the core processing pipelines; otherwise, they remain purely decorative, exposing the system to safety-boundary escapes. Additionally, status validation on high vulnerability must defensively capture any non-modify/non-block status to prevent implicit passes.
**Prevention:** Always implement strict, non-bypassable guardrails in the validator pipeline corresponding to each system degradation mode (e.g., forbidding external tools in Safe Mode, forcing templates in Empathy Override, and enforcing step-by-step confirmation in Conservative Planning).
