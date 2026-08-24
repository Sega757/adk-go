# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.

## 2026-08-08 - [Anti-Caching Security Headers Deployment]
**Vulnerability:** Exposed sensitive session history, execution traces, or intermediate artifact payloads could be aggressively cached by browser history or downstream proxy servers unless anti-caching headers are explicitly configured.
**Learning:** Standard security headers (such as `X-Frame-Options` and `X-Content-Type-Options`) protect against clickjacking and MIME-type sniffing but do not prevent caching. Crucial session and trace endpoints are vulnerable to caching at intermediate proxies or locally on shared workstations unless strict Cache-Control directives are declared globally.
**Prevention:** Always enforce global anti-caching headers (`Cache-Control: no-store, no-cache, must-revalidate, proxy-revalidate`, `Pragma: no-cache`, `Expires: 0`) in all central security middleware of the web and API launchers.
