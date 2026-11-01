## 2026-11-01 - Prevent DoS with http.MaxBytesReader
**Vulnerability:** DoS risk due to using `io.LimitReader` on HTTP requests.
**Learning:** `io.LimitReader` does not close the underlying connection when the read limit is reached, making the server susceptible to resource exhaustion.
**Prevention:** Use `http.MaxBytesReader` for HTTP request bodies as it actively closes the connection and protects against malicious payloads.
