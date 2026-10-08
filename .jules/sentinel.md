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

## 2026-09-03 - [Prevent Stack Trace Leakage in Tool Panic Recovery]
**Vulnerability:** Recovering from panics in function tools (`functiontool`) embedded full stack traces (`debug.Stack()`) in the returned `error` objects, leaking internal call frames, source paths, and function details to external callers, LLM responses, or client payloads.
**Learning:** Error objects returned from tool execution flow back into LLM content or API responses. Exposing stack traces in error strings leaks application internals and creates security risks.
**Prevention:** Log stack traces to server logs using `log.Printf` for operator debugging, and return concise error messages without `debug.Stack()` to the caller.

## 2026-09-06 - [Memory Exhaustion] Prevent DoS via io.LimitReader on HTTP Requests
**Vulnerability:** Found `io.LimitReader` usage on `http.Request.Body` in `server/agentengine/controllers/agent_engine.go`.
**Learning:** `io.LimitReader` does not close the underlying connection when the limit is reached, which exposes the server to DoS attacks by allowing an attacker to send arbitrarily large payloads that exhaust server resources. The server has to keep reading from the connection even when the limit is hit, or close it awkwardly.
**Prevention:** Always wrap `http.Request.Body` with `http.MaxBytesReader` to set a hard limit (e.g. 10MB) before reading or decoding HTTP payload streams. `http.MaxBytesReader` correctly aborts the HTTP request and closes the connection if the limit is exceeded.

## 2026-09-07 - [Prevent Information Leakage in WebSocket Internal Error Close Reason]
**Vulnerability:** In `RunLiveHandler` (`server/adkrest/controllers/runtime.go`), WebSocket close frame reasons for internal server errors (`CloseInternalServerErr` / 1011) sent raw error strings (e.g., `err.Error()`, agent loader failure details) to clients, exposing application internals.
**Learning:** WebSocket close frame reasons are sent directly to connected clients. Sending raw error messages on internal failures (1011) can leak internal file paths, stack traces, or configuration details.
**Prevention:** Log detailed errors server-side using `log.Printf`, and return a sanitized generic reason like `"internal server error"` in WebSocket `CloseInternalServerErr` frames.

## 2026-09-25 - [SSRF bypass via 0.0.0.0 / Unspecified IPs]
**Vulnerability:** The outbound HTTP client wrapper `clientWithSSRFProtection` in `agentregistry` checked for loopback, private, and link-local IPs but failed to block unspecified IP addresses (e.g., `0.0.0.0` or `::` in IPv6).
**Learning:** `0.0.0.0` and `::` (unspecified IPs) are often routed to the local machine (`localhost`) by many operating systems (e.g. Linux/macOS) when passed to socket connect APIs, bypassing basic loopback SSRF checks and allowing attackers to reach internal services.
**Prevention:** Always include `ip.IsUnspecified()` when enforcing SSRF protections using an IP blocklist approach in custom `net.Dialer` contexts.

## 2026-10-05 - [Prevent Information Leakage in REST Controller Error Responses]
**Vulnerability:** In `RunHandler` (`server/adkrest/controllers/runtime.go`), `validateSessionExists` and `decodeRequestBody` returned wrapped raw error messages (e.g., `failed to get session: <err.Error()>` or `failed to decode request: <err.Error()>`) with non-500 status codes (404/400). `NewErrorHandler` preserved non-500 status errors verbatim, leaking internal database errors or JSON decoding details in HTTP response bodies to callers.
**Learning:** `NewErrorHandler` preserves non-500 `statusError` strings. Returning raw wrapped errors (like `fmt.Errorf("failed to get session: %w", err)`) in non-500 status errors directly exposes internal backend error details (such as GORM/database connection failures or internal file paths) to API clients.
**Prevention:** Log detailed error information server-side using `log.Printf` and return sanitized, generic error strings (e.g. `"not found"` or `"bad request"`) in `newStatusError`.

## 2026-10-10 - [Prevent Nil Pointer Dereference DoS in Debug Handlers]
**Vulnerability:** In `DebugAPIController` (`server/adkrest/controllers/debug.go`), `EventSpanHandler` and `SessionSpansHandler` dereferenced `c.debugTelemetry` without checking if it was `nil`.
**Learning:** When optional controller services (such as debug telemetry) are uninitialized or disabled (`nil`), HTTP endpoints attempting to access them will panic on incoming requests, causing a Denial of Service (DoS) vulnerability.
**Prevention:** Always check if optional service dependencies are `nil` before calling methods on them in HTTP handlers, returning an appropriate HTTP response (such as 404 Not Found or an empty dataset) instead of allowing panic.

## 2026-10-15 - [Prevent Information Leakage in SSE Error Events]
**Vulnerability:** In `RunSSEHandler` (`server/adkrest/controllers/runtime.go`), `flashErrorEvent` sent `origError.Error()` directly in SSE stream JSON payloads (`{"error":"..."}`), leaking internal error details without server-side logging.
**Learning:** SSE streaming error helpers must sanitize client error messages to generic strings like `"internal server error"` and log detailed error context server-side with `log.Printf`.
**Prevention:** Always log internal errors server-side before serializing SSE event payloads, and send sanitized generic error messages in client-facing event JSON.
