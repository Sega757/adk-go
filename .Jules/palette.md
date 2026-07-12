# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2025-02-15 - [Interactive CLI Prompt Colorization and Safe TTY-Device Fallbacks]
**Learning:** Adding visual distinction (using emojis and high-contrast colors) in terminal prompts significantly enhances the readability of conversations in terminal-based interactive loops. However, writing ANSI escape sequences to non-TTY targets (like redirected files, logs, or pipes) results in garbled text. Guarding output with safe, TTY-aware checks ensures visual quality without degrading automated execution. Additionally, trimming and ignoring whitespace/empty inputs prevents sending accidental, empty requests to downstream AI services, saving API tokens and avoiding system errors.
**Action:** When developing CLI loops or prompt utilities, check if stdout/stdin is an interactive terminal (`os.ModeCharDevice`). Only inject color codes and emoji graphics if TTY is supported, and always trim inputs before processing to ensure clean navigation.
