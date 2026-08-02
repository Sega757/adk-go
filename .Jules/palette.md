# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-07-26 - TTY-Aware CLI Prompting and API Call Optimization
**Learning:** Reusable UX pattern for CLI design systems involves wrapping terminal prompts with standard-library-only character device checks (`os.ModeCharDevice`). This avoids corrupting output with raw ANSI escape sequences when commands are piped or run in logs, while still presenting rich color-coded prompts (User, Agent, Error, and Human-In-The-Loop/HITL states) to interactive users. Additionally, filtering out empty or whitespace-only inputs at the CLI reader layer prevents expensive, redundant API or LLM queries, keeping the interface snappy and preventing unnecessary cloud costs.
**Action:** Always verify TTY capability using stdlib file descriptor stat checks before emitting CLI styles, and implement robust local input-skipping logic for empty turns in CLI interactive loops.
