# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-24 - CLI Console Launcher Interactive Experience Enhancement
**Learning:** For command-line (CLI) conversational interfaces, distinguishing different speakers (User vs Agent vs System Errors vs Human-in-the-Loop prompts) visually using high-contrast colors and descriptive emojis when connected to an interactive TTY greatly improves readability and context retention. Additionally, ignoring empty or whitespace-only inputs directly in the CLI prevents redundant and costly backend/LLM API calls, making the interaction loop feel smoother and more robust.
**Action:** When designing or refactoring interactive command line prompt loops, check standard output TTY-ness (`os.Stdout.Stat()`) to dynamically apply ANSI escape coloring and emojis, and always perform input sanitization/filtering (such as trimming space and ignoring empty inputs) before invoking the backend runner.

## 2026-03-24 - CLI Help Accessibility and Shortcut Hints
**Learning:** In CLI applications, command listings that omit key accessibility shortcuts (such as standard signal exit combinations like Ctrl+C) or action details increase cognitive load and friction for terminal users. Explicitly pairing slash commands with equivalent system keyboard shortcuts directly in standard and TTY help output improves discoverability and user onboarding.
**Action:** Always include key action descriptions and universal keyboard shortcuts (e.g. Ctrl+C) alongside slash commands in terminal help outputs.

## 2026-03-24 - CLI Confirmation Prompt Discoverability & TTY Color Hierarchy
**Learning:** In interactive CLI applications with confirmation prompts (e.g., human-in-the-loop tool confirmations), stating only one option (such as `Type 'yes' to confirm`) when multiple standard shortcuts are supported (like `y`, `yes`, `confirm`) creates unnecessary user friction and hesitation. Explicitly communicating accepted shortcuts (e.g. `'y'` or `'yes'`) while applying subtle ANSI color highlighting in TTY mode improves user confidence, discoverability, and scanning speed.
**Action:** When writing CLI confirmation prompts, explicitly list common shortcut options (e.g., `'y'` or `'yes'`), apply conditional ANSI formatting when `tty` is true, and maintain clean plaintext output when piped or un-styled.
