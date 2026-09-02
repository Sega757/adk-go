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

## 2026-09-02 - [Upfront Typed Tool Extraction in Flow Loop]
**Vulnerability / Inefficiency:** Converting dynamic request parameters (such as  map of ) to strongly-typed internal interfaces () inside a model response streaming loop caused repeated map allocations and type checks per chunk, and delayed tool validation until LLM responses were processed.
**Learning:** Extracting and validating dynamic tool maps once prior to entering the LLM execution/streaming loop guarantees that tool map contents implement expected interfaces upfront, failing fast on invalid tool configurations before model response streaming starts.
**Prevention:** Always extract and validate dynamic map types upfront into strongly-typed maps once prior to execution loops.

## 2026-09-02 - [Stream Consumer Channel Teardown & Goroutine Leaks]
**Vulnerability / Concurrency Defect:** Unbuffered channel sends on  or  within streaming background goroutines in  lacked  guards against context cancellation (). If a consumer stopped listening mid-stream or encountered an error, the background producer goroutine would block indefinitely, leaking memory and connections.
**Learning:** Background producer goroutines writing stream events or errors onto channels must always write using  with  or  to guarantee non-blocking exit when consumers cancel or fail.
**Prevention:** Always pair unbuffered channel sends in streaming producer goroutines with  listening on context cancellation ().

## 2026-09-02 - [Tool Execution Timeout Enforcement]
**Vulnerability / Flow Stall:** Unbounded or slow third-party tool execution could hang multi-step agent flow execution indefinitely if a tool failed to respond or blocked on I/O.
**Learning:** Configured  in  must be enforced per tool call task using a child context (). If  is unset or zero, the parent context deadline is preserved as the fallback (rather than forcing an arbitrary default timeout), and timeout expiration is returned as an error response payload event () so the model can inspect and recover gracefully.
**Prevention:** Always bound tool task execution using child contexts derived from  (falling back to parent context deadline when zero), capturing deadline expiration as structured tool error payloads.

## 2026-09-02 - [Panic-Free Conformance Loader & Node Registration]
**Vulnerability / Panic Exposure:** Registering custom node functions or initializing conformance loaders with empty names, nil function pointers, or duplicate identifiers previously caused panics instead of returning Go  types, exposing configuration/loader callers to runtime crashes.
**Learning:** Loader utilities and registry functions (, , ) must perform pre-lock input validation (checking for empty names,  function pointers,  agent map entries) and return explicit, descriptive Go  values rather than panicking.
**Prevention:** Always validate loader inputs and registration parameters upfront and propagate structured Go  returns instead of invoking .
