# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2025-02-15 - Preventing Redundant LLM Overhead in CLI Shells
**Learning:** In interactive CLI agent prompts, users frequently press 'Enter' accidentally or submit empty/whitespace-only lines. If not handled at the interface boundary, these empty inputs trigger expensive, slow, and redundant network/LLM API calls, degrading the perceived performance and responsiveness of the application.
**Action:** Always intercept empty or whitespace-only inputs at the CLI input loop before dispatching runner/model execution. Provide instant terminal feedback (re-printing the prompt) rather than silently waiting on a backend timeout.
