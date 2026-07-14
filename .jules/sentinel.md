# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-14 - Dynamic CORS Origin Validation and Cache Poisoning Prevention
**Vulnerability:** Insecure CORS policy and lack of dynamic Origin header validation in the web REST API allowed potential cross-origin bypasses or invalid scheme-less address mismatches, combined with missing HTTP Vary headers that could lead to cache poisoning.
**Learning:** Statically setting `Access-Control-Allow-Origin` to a configured value that lacks a scheme (like "localhost:8080") causes browsers to reject cross-origin requests. Simply reflecting the incoming origin header is highly vulnerable to suffix matching bypasses unless exact string matching (with and without schemas) is strictly enforced. Furthermore, dynamically modifying the CORS response headers without declaring `Vary: Origin` exposes downstream caches to serving incorrect Access-Control headers to mismatched origins.
**Prevention:** Always validate incoming `Origin` headers exactly (supporting both schemed and scheme-less target configurations) rather than relying on weak regular expressions or substring lookups. Explicitly append `Vary: Origin` whenever dynamic CORS response headers are generated, and fail securely with a `403 Forbidden` response when an invalid origin attempts to connect.
