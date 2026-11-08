## 2026-11-08 - DoS Vulnerability via `io.LimitReader`
**Vulnerability:** Using `io.LimitReader` on `http.Request.Body` does not close the underlying connection when the limit is reached, exposing the server to Denial of Service (DoS) attacks through resource exhaustion.
**Learning:** We assumed `io.LimitReader` safely bounded input reads. However, because it doesn't close the connection, an attacker can hold the connection open or we leak resources if they keep sending data.
**Prevention:** Always use `http.MaxBytesReader` for bounding reads from HTTP request bodies in handlers, as it correctly aborts the request and closes the connection if the limit is exceeded.
