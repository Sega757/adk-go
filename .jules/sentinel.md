# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-27 - [Global Anti-Caching and AgentEngine Security Headers]
**Vulnerability:** Browser caching of sensitive REST/AgentEngine API endpoints and lack of standard secure headers on the AgentEngine HTTP handler. This could lead to information exposure if local browsers or intermediate proxies cache sensitive session traces, history, or responses, and clickjacking/MIME-sniffing/XSS on AgentEngine endpoints.
**Learning:** Standard REST API, web launcher, and AgentEngine handlers must enforce anti-caching headers (`Cache-Control: no-store, no-cache, must-revalidate, proxy-revalidate`, `Pragma: no-cache`, `Expires: 0`) and standard security headers globally.
**Prevention:** Register standard HTTP security and anti-caching middleware globally on all mux routers serving REST/AgentEngine endpoints, and directly verify their enforcement using package-level HTTP test recorders in the test suites.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.
