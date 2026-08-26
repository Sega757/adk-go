# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-08-15 - [Load Artifacts Tool Input Validation]
**Vulnerability:** Unsanitized artifact names supplied by model function calls or tool arguments in `loadartifactstool` could contain path separators (`/`, `\`) or path traversal sequences (`..`).
**Learning:** Tools that interface between LLM function calls/responses and underlying context/services must validate user/model inputs at the tool boundary, even if downstream services also perform validation.
**Prevention:** Always validate tool string inputs using dedicated validator helpers (e.g., checking for `/`, `\`, and `..`) before processing or invoking underlying service calls.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.
