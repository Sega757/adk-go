# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-29 - Stylized CLI Prompts & TTY Heuristics
**Learning:** Adding ANSI escape codes and emojis to terminal commands significantly increases interactive visual delight and structural clarity. However, hardcoding ANSI escapes breaks pipeline scripts and produces corrupted plaintext/log outputs. Checking TTY status using stdlib-only file descriptor stats (`os.Stdout.Stat()`) avoids introducing external dependencies while ensuring clean non-interactive fallbacks.
**Action:** Always wrap stylized console output behind an `isTTY()`-style check, and use robust channels with explicit closure (`defer close`) to handle unexpected standard input interruptions gracefully.
