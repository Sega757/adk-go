# Palette's UX and Accessibility Journal

This journal documents critical UX and accessibility learnings, patterns, and insights discovered during the enhancement of the Agent Development Kit (ADK) codebase.

---

## 2026-07-12 - Defensive EOF Channel Handling in CLI Input Loops
**Learning:** When optimizing a Go CLI application to ignore empty/whitespace inputs (avoiding redundant backend/model requests), checking input without the comma-ok idiom can introduce a critical bug where EOF causes an infinite loop of empty reads, spiking CPU usage to 100%. If a reader goroutine finishes on EOF, it must close the input channel, and the receiving loop must handle the closed channel state using the comma-ok pattern.
**Action:** Always pair empty-input checks in Go CLI select loops with a check on the channel closed status (`case val, ok := <-chan`) and defer a channel close on the reader goroutine to gracefully exit.
