# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-02-15 - Dynamic CORS Validation with Cache-Poisoning Mitigation
**Vulnerability:** The CORS middleware (`corsWithArgs`) blindly trusted and set the configured frontend address in `Access-Control-Allow-Origin` without validating the incoming request's `Origin` header. It also lacked the `Vary: Origin` header, creating a risk of HTTP cache poisoning, and failed to reject unauthorized cross-origin requests.
**Learning:** Hardcoding or setting a single static string in `Access-Control-Allow-Origin` is insufficient and breaks when both schemed/scheme-less addresses are supported. Dynamic origin matching with standard schemes (`http://` and `https://`) must be carefully evaluated for incoming `Origin` headers. Additionally, failing to reject unmatched origins allowed unauthorized CORS requests to hit core API handlers before CORS could be checked by browsers.
**Prevention:** Always dynamically match incoming request `Origin` headers against a normalized allowed origin list. Securely reject unmatched origins with a `403 Forbidden` response to fail-secure. Always set `Vary: Origin` whenever CORS headers are returned dynamically.
