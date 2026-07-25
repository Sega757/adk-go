# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-05-15 - [Dynamic CORS Origin Verification]
**Vulnerability:** A static or wildcard CORS origin configuration allowed any incoming HTTP request to either bypass origin checks or cause incorrect CORS header generation, potentially exposing sensitive backend APIs to cross-origin data theft (such as CSRF or cross-origin read vulnerabilities).
**Learning:** Naive middleware implementation blindly trusted the configured `webui_address` without dynamically checking incoming requests' `Origin` headers, and without securely rejecting requests with unrecognized origins.
**Prevention:** Always parse and validate request `Origin` headers dynamically against configured white-lists (handling both schemed and scheme-less formats correctly), set the `Vary: Origin` header to prevent cache poisoning, and reject unauthorized origins with a strict 403 Forbidden status code.
