# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-10 - CLI Terminal Rich Feedback and Graceful Fallback
**Learning:** Terminal-based user interfaces (CLIs) often struggle with readability when displaying continuous agent-human interactions. Using ANSI colors and emojis significantly improves the scanning of prompt states (User/Agent/Error), but can cause control character garbage/noise in non-interactive (non-TTY) or piped environments. A robust pattern checks TTY-ness using `os.Stdout.Stat()` (stdlib-only, no external dependencies) to dynamically toggle rich styling vs plaintext.
**Action:** Always check TTY capabilities before emitting colored/formatted terminal outputs, and maintain a seamless, legible plaintext fallback for pipelines and automated environments.
