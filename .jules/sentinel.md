# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-02-18 - Dynamic CORS Origin Validation and Cache Poisoning Prevention
**Vulnerability:** Permissive or standard configuration-based CORS middleware (`Access-Control-Allow-Origin`) can lead to cross-origin bypasses, cross-site request forgery (CSRF), and cache poisoning if the user-configured `webui_address` (scheme-less or schemed) is output directly without validating against incoming request `Origin` headers, or if intermediate caches cache the wildcard or single origin response for all origins.
**Learning:** Browsers require exact matches of scheme, host, and port for CORS. Directly using a scheme-less address like `localhost:8080` in `Access-Control-Allow-Origin` causes standard browsers to reject CORS requests, while blindly accepting any origin can leak sensitive session information. Properly parsing the configured backend address, comparing the incoming `Origin` scheme and host, returning `403 Forbidden` for invalid or unmatched origins, and attaching `Vary: Origin` ensures both compliance and high security.
**Prevention:** Always implement dynamic validation of `Origin` headers against configured schemed or scheme-less hostnames in custom Go CORS middlewares, enforce strict scheme matching when schemes are specified to prevent cross-scheme CORS bypasses, and append `Vary: Origin` to mitigate intermediate proxy caching issues.
