# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-03-01 - CLI Terminal High-Contrast Color and Emoji-Enhanced Prompts
**Learning:** When building console/terminal interfaces, users greatly benefit from high-contrast visual demarcations and context markers (such as emojis or colored labels) to distinguish between their inputs, agent responses, and error conditions. Ensuring these enhancements are safe (falling back to plaintext gracefully when run outside of a TTY) guarantees universal accessibility.
**Action:** Use a dedicated `isTTY()` check before applying ANSI styles or complex emojis in terminal commands, and define robust formatting helpers to standardise prompts across multi-turn human-in-the-loop (HITL) processes.
