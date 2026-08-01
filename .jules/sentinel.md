# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-21 - Dynamic CORS Origin Validation and Cache Poisoning Prevention
**Vulnerability:** Insecure CORS implementation where `Access-Control-Allow-Origin` was hardcoded to a configured web address. If the configured frontend address was scheme-less (like `localhost:8080`), standard browsers rejected the header as invalid. Reflecting the value without checking incoming request origins also risked unauthorized cross-origin requests.
**Learning:** Modern browsers require exact CORS origin matches, including scheme, host, and port. Directly reflecting the header dynamically requires strict host and scheme checks to prevent unauthorized access, and appending `Vary: Origin` prevents intermediate caches from poisoning responses.
**Prevention:** Always dynamically parse and validate incoming request `Origin` headers against the allowed configuration, enforce matching schemes if configured to prevent cross-scheme CORS bypasses, reject mismatches with 403 Forbidden, and append `Vary: Origin` to secure proxy caching.
