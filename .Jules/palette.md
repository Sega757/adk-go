# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2025-03-01 - [Terminal CLI Interactive UX Pattern]
**Learning:** For terminal-based CLI interactive environments, user experience is dramatically improved by recognizing stdout TTY capabilities. This allows us to conditionally render high-contrast greeting banners, and distinct color-coded, emoji-enhanced prompts for human vs. AI roles (reducing visual clutter and helping users instantly parse who is talking). Furthermore, handling input whitespace-only checks client-side prevents redundant, slow network roundtrips to backends/APIs.
**Action:** Always inspect terminal output capability (`os.Stdout` stats for character device on Go) to dynamically scale visual richness. Pair this with robust EOF channel closures and input pruning to guarantee a responsive and reliable console experience.
