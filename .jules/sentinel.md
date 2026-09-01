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

## 2026-09-01 - [Missing ReadHeaderTimeout Config]
**Vulnerability:** The HTTP server implementation in `cmd/launcher/web/web.go` was previously not explicitly configuring `ReadHeaderTimeout`.
**Learning:** A missing `ReadHeaderTimeout` in `http.Server` makes the server vulnerable to Slowloris-style denial-of-service (DoS) attacks, in which an attacker sends request headers very slowly, keeping the connection open and exhausting server resources.
**Prevention:** Ensured `ReadHeaderTimeout` is exposed via a configurable command-line flag (`-read-header-timeout`, with a safe default of `5s`) and explicitly applied to `http.Server` initialization.

## 2026-09-01 - [Consistent Logging in Instruction Template Processing]
**Vulnerability/Issue:** Silent failure during instruction placeholder variable injection for optional artifacts and optional session state keys made debugging missing state/artifacts difficult without surfacing errors to the LLM flow.
**Learning:** Optional template variables (`{artifact.foo?}` or `{bar?}`) should gracefully resolve to empty strings when missing or when benign lookup errors occur, but unexpected errors (such as artifact loading failures or session state query errors other than `session.ErrStateKeyNotExist`) must be logged for operator visibility.
**Prevention:** Always use standard `log.Printf` to log unexpected optional lookup failures while preserving fallback behavior (`return "", nil`), ensuring `session.ErrStateKeyNotExist` remains silently ignored as expected missing state.
