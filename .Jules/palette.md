# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-07-17 - Clear CLI Onboarding with Visually Distinct Welcome Banners
**Learning:** CLI terminal utilities can be confusing or feel unguided for first-time users. Providing a beautifully color-coded, emoji-enriched welcome banner when run inside a TTY, coupled with a simple text-only fallback, significantly increases usability and decreases onboarding friction without altering any functional logic.
**Action:** When designing interactive console/CLI tools, always present a polite and structured welcome header with explicit instructions on how to interact (e.g., input confirmation keys) and clear directions on how to gracefully terminate the session (e.g., Ctrl+D or Ctrl+C shortcuts).
