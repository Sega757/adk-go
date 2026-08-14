# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-24 - CLI Console Launcher Interactive Experience Enhancement
**Learning:** For command-line (CLI) conversational interfaces, distinguishing different speakers (User vs Agent vs System Errors vs Human-in-the-Loop prompts) visually using high-contrast colors and descriptive emojis when connected to an interactive TTY greatly improves readability and context retention. Additionally, ignoring empty or whitespace-only inputs directly in the CLI prevents redundant and costly backend/LLM API calls, making the interaction loop feel smoother and more robust.
**Action:** When designing or refactoring interactive command line prompt loops, check standard output TTY-ness (`os.Stdout.Stat()`) to dynamically apply ANSI escape coloring and emojis, and always perform input sanitization/filtering (such as trimming space and ignoring empty inputs) before invoking the backend runner.

## 2026-03-25 - CLI Console Launcher Slash-Command Design Constraints
**Learning:** When adding terminal slash-commands (`/help`, `/clear`, `/exit`, `/quit`) to an interactive CLI agent prompt, intercepting all slash-prefixed inputs aggressively causes a severe UX regression because users can no longer write literal slash-prefixed inputs (like absolute file paths, e.g. `/tmp/file` or `/usr/bin`).
**Action:** When parsing slash commands in interactive terminals, exclusively match the exact registered command names (along with invalid option validations) and immediately pass through any unrecognized slash-prefixed inputs directly to the underlying agent. This ensures maximum convenience for the user while preventing regressions on literal text input.
