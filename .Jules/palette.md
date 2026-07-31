# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-24 - High-Contrast CLI Prompts and TTY Fallbacks
**Learning:** For command-line (CLI) developer interfaces, visual hierarchy and state representation (e.g. User, Agent, Errors, and Human-in-the-Loop prompts) can be significantly improved using ANSI escape colors and emojis. However, to avoid visual "garbage characters" in redirected environments, logging systems, or non-interactive terminals, all such enhancements must detect the terminal's character device status (TTY-ness) and fall back to clean, plain text.
**Action:** Always implement a zero-dependency `isStdoutTTY()` check via `os.Stdout.Stat()` to toggle between colorized, rich CLI prompts and plain text fallbacks.
