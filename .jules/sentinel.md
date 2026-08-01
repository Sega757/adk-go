# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-02-18 - [Insecure CORS Origin Reflection and Web Cache Poisoning]
**Vulnerability:** The CORS middleware (`corsWithArgs`) previously set the `Access-Control-Allow-Origin` header directly to the unvalidated, user-configured `webui_address`. When configured without a scheme (e.g. `localhost:8080`), it resulted in an invalid CORS header format rejected by browsers. Furthermore, dynamically reflecting configurations/origins without setting the `Vary: Origin` header could lead to web cache poisoning, where cached responses with CORS headers authorized for one origin are erroneously served to another origin.
**Learning:** Hardcoded or raw configuration reflections without strict parsing/scheme normalization and validation against actual request headers create brittle CORS policies and cross-origin access gaps.
**Prevention:** Always validate incoming request `Origin` headers dynamically against a list of normalized, permitted origins (with schemes). When dynamically returning origins in `Access-Control-Allow-Origin`, always set the `Vary: Origin` header to instruct caches to separate cache buckets by request origin.
