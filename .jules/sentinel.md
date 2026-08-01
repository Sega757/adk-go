# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-19 - Insecure CORS Origin Wildcard/Reflection Bypass
**Vulnerability:** The API server had a naive CORS implementation that blindly returned any configured `webui_address` without validating the incoming request's `Origin` header. This could lead to CORS bypass or connection issues when clients connected with mismatching schemes (e.g. HTTP vs HTTPS) or disallowed domains, potentially exposing session tokens/credentials.
**Learning:** Hardcoded or unvalidated CORS response headers in Go can fail to enforce security boundaries across varying transport schemes or domains. Dynamic origin validation, strict scheme matching, and adding `Vary: Origin` headers are essential to protect the REST API endpoints.
**Prevention:** Always validate the incoming `Origin` header dynamically against parsed and normalized allowed schemes and hosts, and reject mismatched origins with 403 Forbidden. Set the `Vary: Origin` header to prevent cache poisoning.
