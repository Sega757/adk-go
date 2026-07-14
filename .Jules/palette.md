# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2025-02-14 - TTY-Sensitive Color Prompts and Empty Input Control in CLI
**Learning:** Terminal CLI users frequently press 'Enter' accidentally or input blank spaces. Submitting these directly to LLM backends is highly inefficient and creates a bad UX. Additionally, printing ANSI escape color codes should only occur if the output stream is an interactive terminal (character device), otherwise it corrupts standard streams or redirected log files.
**Action:** When designing interactive CLI loops, always validate and trim user inputs before passing them to the main runner logic. Use `os.Stdout.Stat()` to check `os.ModeCharDevice` so that color codes and emojis are applied dynamically and safely degraded in non-interactive environments.
