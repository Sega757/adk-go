# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.

## 2026-08-21 - [Unbounded HTTP Request Body Consumption in Eventarc Trigger]
**Vulnerability:** In binary mode HTTP webhook endpoints (such as `EventarcTriggerHandler`), calling `io.ReadAll(r.Body)` directly without a size cap allows malicious clients to stream unlimited payload bytes into server RAM, leading to memory exhaustion and Denial-of-Service (DoS).
**Learning:** `json.NewDecoder` streams JSON token-by-token or returns decoding errors on invalid structure, but `io.ReadAll` blindly buffers all incoming bytes directly into a `[]byte` slice until EOF.
**Prevention:** Always wrap raw `r.Body` reader streams with `http.MaxBytesReader(w, r.Body, maxBytes)` before reading binary payloads with `io.ReadAll`.
