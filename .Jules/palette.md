# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-24 - CLI Console Launcher Interactive Experience Enhancement
**Learning:** For command-line (CLI) conversational interfaces, distinguishing different speakers (User vs Agent vs System Errors vs Human-in-the-Loop prompts) visually using high-contrast colors and descriptive emojis when connected to an interactive TTY greatly improves readability and context retention. Additionally, ignoring empty or whitespace-only inputs directly in the CLI prevents redundant and costly backend/LLM API calls, making the interaction loop feel smoother and more robust.
**Action:** When designing or refactoring interactive command line prompt loops, check standard output TTY-ness (`os.Stdout.Stat()`) to dynamically apply ANSI escape coloring and emojis, and always perform input sanitization/filtering (such as trimming space and ignoring empty inputs) before invoking the backend runner.

## 2026-03-25 - CLI REPL Command-Driven Interactive Control
**Learning:** In interactive CLI applications, allowing users to navigate, document, and clear their sessions via dedicated, keyboard-accessible slash-commands (like `/exit`, `/quit`, `/clear`, `/help`) offers a much more elegant, intuitive, and safe user experience than relying solely on raw keyboard interrupts (like `Ctrl+C`). It prevents unintended shell termination, supports TTY detection for modern high-contrast formatting while maintaining perfect non-TTY plaintext fallbacks for standard automation pipes, and provides immediate help documentation on demand.
**Action:** For all REPL and interactive command line loops, include an entry banner instructing the user on how to get help and quit cleanly, and implement slash-command handlers to intercept commands before sending inputs to downstream processing/LLMs.
