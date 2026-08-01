# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2025-02-15 - Interactive CLI Console Design and Terminal Compatibility
**Learning:** Adding rich visual cues (like color-coding and emojis) to CLI applications greatly improves readability and user engagement, but can break in non-interactive shells, PTYs, or pipes. Checking terminal capabilities (such as TTY device type detection) and offering a clean plaintext fallback ensures standard terminal accessibility without generating visual noise (e.g., ANSI escape sequence clutter) in logs or automated runner streams.
**Action:** Always verify if stdout/stderr is a TTY using standard-library term capabilities (e.g., `os.ModeCharDevice` check) before outputting ANSI colors or interactive emojis in terminal interfaces.
