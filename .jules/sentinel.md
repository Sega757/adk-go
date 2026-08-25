# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.

## $(date +%Y-%m-%d) - [MEDIUM] Missing Input Body Limits
**Vulnerability:** Found multiple web handlers (`CreateSessionHandler` in `server/adkrest/controllers/sessions.go` and `decodeRequestBody` in `server/adkrest/controllers/runtime.go`) accepting JSON bodies without limiting the request size. This could lead to a memory exhaustion Denial-of-Service (DoS) attack if large payloads are sent.
**Learning:** `json.NewDecoder` decodes directly from the provided `io.Reader`. If `req.Body` is unchecked, a malicious payload could stream directly into allocated memory, resulting in server panic due to Out Of Memory conditions.
**Prevention:** In Go REST web servers using `net/http`, standard practice dictates wrapping `req.Body` with `http.MaxBytesReader(rw, req.Body, maxBytes)` before performing JSON decoding or body reads to safely truncate malicious payloads and close connections cleanly.
