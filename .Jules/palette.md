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

## 2026-03-25 - Invalid CLI Command Feedback & Prompt Safeguards
**Learning:** In interactive CLI agent prompts, mistyped slash commands (e.g. `/hepl`, `/foo`) sent directly to backend LLMs waste latency and create confusing LLM responses. Intercepting unrecognized single-word slash commands (excluding path inputs with slash separators like `/tmp/file`) locally, providing immediate color-coded hint feedback, and re-prompting the user avoids unnecessary agent processing and improves CLI feedback loops.
**Action:** When processing interactive command line input loops, intercept single-token slash prefixes locally, surface clear helper guidance, and prevent invalid slash command typos from being sent to backend agent handlers.

## 2026-09-01 - Web UI Chat Stream Accessibility & Focus-Visible States
**Learning:** Live-updating streaming containers (chat messages, event logs) that omit `role="log"` and `aria-live="polite"` prevent screen readers from dynamically announcing updates. Pairing live region roles with explicit `aria-label`s on controls and explicit CSS `:focus-visible` outline rules for interactive buttons ensures both screen reader users and keyboard navigators receive immediate focus and feedback cues.
**Action:** Always assign `role="log"` and `aria-live="polite"` to dynamically updated message/console logs and define CSS `:focus-visible` indicators for interactive elements.

## 2026-09-02 - Web UI Streaming Toggle Accessibility & Active Recording Feedback
**Learning:** For interactive media streaming controls (like voice input or camera video toggles), omission of `aria-pressed` prevents screen reader users from identifying whether recording/streaming is currently live. Updating `aria-pressed` dynamically in JavaScript alongside CSS keyframe pulse animations on active `.active` toggle buttons provides unambiguous visual and assistive technology feedback during live recording sessions.
**Action:** Always include initial `aria-pressed="false"` on stateful toggle buttons, sync `aria-pressed` dynamically in toggle handlers, and pair active states with high-contrast visual indicators such as subtle keyframe pulse shadows.

## 2026-09-03 - Web Modal Focus Management & Dynamic Input State Safeguards
**Learning:** For modal dialogs in streaming web apps, opening a modal without shifting keyboard focus or trapping Tab navigation leaves screen reader and keyboard users interacting with disabled background controls. Shifting focus to the primary modal action on open, trapping Tab navigation within modal bounds, restoring focus to the triggering element on close, and synchronizing input `disabled` attributes with connection status ensures robust accessibility and state clarity.
**Action:** Always store `previouslyFocusedElement` when opening modals, focus the primary modal action button, constrain Tab navigation to modal focusables, restore focus on close, and disable text inputs when backend connections are inactive.
