# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-04 - High-Contrast CLI Prompts and Graceful Empty Input Handling in Go
**Learning:** Interactive command line interfaces benefit tremendously from color-coded, emoji-enriched visual cues to distinguish between user actions, agent responses, errors, and Human-in-the-Loop (HITL) prompt requests. However, such styling must always be guarded by TTY detection (e.g. `fi.Mode()&os.ModeCharDevice != 0`) to prevent raw ANSI escape codes from cluttering redirected log files or piping flows. Furthermore, safely ignoring empty/whitespace-only input locally prevents wasteful and redundant API calls.
**Action:** Always check TTY status before emitting ANSI codes/emojis in terminal-based interfaces, use yellow/amber headers for interrupts/HITL to grab attention, and implement client-side/inline empty input checks to shield upstream APIs from redundant processing.
