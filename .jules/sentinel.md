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

## 2024-03-22 - [Memory Exhaustion] Prevent DoS via Unbounded Payloads
**Vulnerability:** Found unconstrained `io.ReadAll` and `json.NewDecoder` usage on `http.Request.Body` in trigger handlers (Eventarc & PubSub).
**Learning:** This exposes the server to DoS attacks by allowing an attacker to send arbitrarily large payloads that exhaust server memory before the request can be processed.
**Prevention:** Always wrap `http.Request.Body` with `http.MaxBytesReader` to set a hard limit (e.g. 10MB) before reading or decoding HTTP payload streams.

## 2026-08-29 - [Missing HTTP Server Timeouts]
**Vulnerability:** Found `http.ListenAndServe` in REST API example which launches a server without explicit timeouts.
**Learning:** By default, Go's `net/http` package does not set timeouts for reading headers, reading the body, or writing responses. This exposes the server to slow-client Denial of Service (DoS) attacks, such as Slowloris, where an attacker intentionally sends data very slowly to exhaust the server's connection pool.
**Prevention:** Always instantiate an explicit `http.Server` and set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` fields before calling `ListenAndServe()`.
## 2026-08-31 - [Configurable Tool Reference Type Safety & Registry Isolation]
**Vulnerability/Risk:** In `internal/configurable/configurable_utils.go`, looking up entries in `toolRegistry` that do not implement `ToolFactory` or `ToolsetFactory` could cause unexpected nil pointer/interface handling or runtime errors if type assertions fail unhandled. Additionally, tests modifying `toolRegistry` concurrently or sequentially without cleanup risk polluting global state.
**Learning:** `ResolveToolReference` explicitly checks type assertions (`ToolFactory` and `ToolsetFactory`) and returns a formatted error on type mismatches. Unit tests targeting registry references must isolate test state using `resetRegistries(t)` with `t.Cleanup` to restore original registry contents upon test completion.
**Prevention:** Always cover type assertion fallback branches in unit tests with invalid dummy entries, and enforce `resetRegistries(t)` state isolation pattern in tests that register items in global package-level maps.
