# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-05 - Safe Terminal-Only Interactive Polish for LLM REPLs
**Learning:** For command-line interfaces interacting with slow LLM APIs, users frequently press "Enter" on empty or whitespace-only inputs, which traditionally triggers a redundant, expensive network request. Local validation that intercepts empty/whitespace input or graceful exit commands (like `/exit`, `/quit`) keeps the interface snappy and saves resources. Furthermore, colorized ANSI themes and rich Unicode/emojis greatly enhance the visual hierarchy (separating user from model) but must dynamically fallback to raw plaintext when standard output is redirected or non-TTY (using `fi.Mode() & os.ModeCharDevice != 0`) to prevent visual noise or corruption in log files and pipes without introducing external term/color dependencies.
**Action:** In CLI REPL loops, always intercept empty/whitespace inputs locally to avoid redundant API requests, handle `/exit` and `/quit` slash commands, and use standard-library TTY checks to apply ANSI themes conditionally.
