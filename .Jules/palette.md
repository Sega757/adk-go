# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-07-26 - TTY-Aware High-Contrast Visual Framing in CLI REPL Interfaces
**Learning:** Users interacting with terminal/CLI interfaces benefit immensely from visual structure and color/icon-based distinction between roles (User vs. Agent), but raw ANSI escape sequences must never pollute automated or non-interactive stdout (such as redirected files, pipes, or testing mock run buffers).
**Action:** Use a standard-library-only `os.ModeCharDevice` mode check on standard output to dynamically determine TTY-ness, cleanly toggling between high-contrast rich formatting and standard plaintext fallbacks.
