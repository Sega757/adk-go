# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-24 - CLI Console Launcher Interactive Experience Enhancement
**Learning:** For command-line (CLI) conversational interfaces, distinguishing different speakers (User vs Agent vs System Errors vs Human-in-the-Loop prompts) visually using high-contrast colors and descriptive emojis when connected to an interactive TTY greatly improves readability and context retention. Additionally, ignoring empty or whitespace-only inputs directly in the CLI prevents redundant and costly backend/LLM API calls, making the interaction loop feel smoother and more robust.
**Action:** When designing or refactoring interactive command line prompt loops, check standard output TTY-ness (`os.Stdout.Stat()`) to dynamically apply ANSI escape coloring and emojis, and always perform input sanitization/filtering (such as trimming space and ignoring empty inputs) before invoking the backend runner.

## 2026-03-25 - Terminal REPL Slash-Commands Navigation & Safety Safeguards
**Learning:** For command-line interfaces (CLI) running interactive LLM REPL loops, intercepting special control commands (like `/exit`, `/quit`, `/help`, `/clear`) locally on the client-side before sending the input to the agent or LLM prevents unnecessary LLM token generation, prevents weird LLM responses to exit desires, and guides users on how to control their terminal sessions. Implementing dynamic TTY styling with fallback plaintext ensures these commands are fully accessible and render correctly regardless of terminal configuration.
**Action:** Always capture client-side slash-commands directly in the console input loop before passing user queries downstream to the model runner, and format the output according to the terminal TTY-ness.
