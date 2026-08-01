# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-03-05 - CORS Origin Validation Bypass & Cache Poisoning Mitigation
**Vulnerability:** Static or permissive CORS setups that directly output a configured hostname as `Access-Control-Allow-Origin` (or echo the input without validation) can cause browser CORS rejections or lead to cross-site request hijacking. If dynamic CORS returns the request's origin without `Vary: Origin`, it exposes responses to intermediate and browser cache poisoning.
**Learning:** The previous ADK REST API CORS middleware set `Access-Control-Allow-Origin` to the exact static `webui_address` configuration value directly, which defaults to a scheme-less value (`localhost:8080`). This was invalid under the browser CORS specification and lacked dynamic origin checking, scheme matching, or cache-poisoning mitigations.
**Prevention:** Implement a dynamic CORS validation middleware that:
1. Dynamically parses the request's `Origin` header.
2. Validates the incoming origin's scheme, host, and port against the allowed configuration (handling both schemed and scheme-less formats correctly and preventing cross-scheme CORS bypass).
3. Adds `Vary: Origin` to ensure intermediate caches separate responses based on the request origin.
4. Securely rejects unauthorized origins with `403 Forbidden`.
