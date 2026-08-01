# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-02-14 - Robust CORS Origin Validation and Varying
**Vulnerability:** The REST API CORS middleware previously used a static `frontendAddress` (which could be scheme-less like `localhost:8080`) for the `Access-Control-Allow-Origin` header without validating the incoming request's `Origin` header. This could allow cache-poisoning (when a response is cached for a different origin) or permit cross-scheme CORS bypasses.
**Learning:** Setting dynamic `Access-Control-Allow-Origin` headers safely requires robust parsing of both the incoming `Origin` and the configured allowed host. If the configured host does not specify a scheme, we should match any scheme (e.g. `http` or `https`) but still perform host and port verification. If the configured address contains a scheme, we must strictly enforce the scheme to prevent cross-scheme bypasses. To prevent browser and proxy cache poisoning, the `Vary: Origin` header must be applied.
**Prevention:** Always parse incoming `Origin` headers as URLs, extract hostnames and schemes cleanly, validate them dynamically against authorized hosts, use `Vary: Origin`, and reject invalid or unmatched origins with `403 Forbidden` before setting any Access-Control headers.
