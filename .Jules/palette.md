# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-24 - CLI Console Launcher Interactive Experience Enhancement
**Learning:** For command-line (CLI) conversational interfaces, distinguishing different speakers (User vs Agent vs System Errors vs Human-in-the-Loop prompts) visually using high-contrast colors and descriptive emojis when connected to an interactive TTY greatly improves readability and context retention. Additionally, ignoring empty or whitespace-only inputs directly in the CLI prevents redundant and costly backend/LLM API calls, making the interaction loop feel smoother and more robust.
**Action:** When designing or refactoring interactive command line prompt loops, check standard output TTY-ness (`os.Stdout.Stat()`) to dynamically apply ANSI escape coloring and emojis, and always perform input sanitization/filtering (such as trimming space and ignoring empty inputs) before invoking the backend runner.

## 2026-03-25 - Interactive Session Management with Console Slash Commands
**Learning:** Adding dedicated, discoverable slash-prefix commands (like `/help`, `/clear`, `/exit`, `/quit`) to interactive command-line loops provides an intuitive, accessible alternative to lower-level OS interrupt signals (such as Ctrl+C or EOF / Ctrl+D). This improves the user experience by giving clear exit strategies and screen management utilities within the conversational session context.
**Action:** Include command routing logic inside text prompt loops to intercept prefix-based commands, cleanly handling terminal buffer resets and graceful session teardowns before sending messages to backend services.
