# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-24 - CLI Console Launcher Interactive Experience Enhancement
**Learning:** For command-line (CLI) conversational interfaces, distinguishing different speakers (User vs Agent vs System Errors vs Human-in-the-Loop prompts) visually using high-contrast colors and descriptive emojis when connected to an interactive TTY greatly improves readability and context retention. Additionally, ignoring empty or whitespace-only inputs directly in the CLI prevents redundant and costly backend/LLM API calls, making the interaction loop feel smoother and more robust.
**Action:** When designing or refactoring interactive command line prompt loops, check standard output TTY-ness (`os.Stdout.Stat()`) to dynamically apply ANSI escape coloring and emojis, and always perform input sanitization/filtering (such as trimming space and ignoring empty inputs) before invoking the backend runner.

## 2026-03-25 - CLI Console Session Slash-Commands Integration
**Learning:** For terminal-based REPL or conversational agents, offering structured interactive commands (such as `/help`, `/clear`, `/exit`, `/quit`) rather than expecting users to know terminal signals (e.g. Ctrl+C) or raw commands significantly reduces friction, increases accessibility, and prevents session control errors. It provides an intuitive, discoverable, and standard interface.
**Action:** Always intercept user text inputs for slash prefixes (`/`) and route them to structured local handler routines (help guides, visual clear commands, graceful shutdowns) before processing them as typical conversational inputs.
