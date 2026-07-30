# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-05-10 - Enforcing Global Anti-Caching Security Headers for Rest API & Web Launcher
**Vulnerability:** Downstream web proxies, reverse proxies, and browsers could cache sensitive dynamic responses (such as execution traces, model responses, system logs, or session secrets) if anti-caching headers are omitted.
**Learning:** Even if custom security headers (like X-Frame-Options, X-Content-Type-Options) are applied, caching is governed independently by HTTP cache control policies. The REST API backend and console/web servers must actively reject intermediate caching.
**Prevention:** Always bundle anti-caching directives (`Cache-Control: no-store, no-cache, must-revalidate, proxy-revalidate`, `Pragma: no-cache`, `Expires: 0`) inside the global security headers middleware to guarantee defense-in-depth on all dynamic endpoints.
