# Palette UX & Accessibility Ledger

## 2026-11-09 - Accessible Modal Dialog & Focus Management

**Learning:**
Custom modals in static web UI examples (such as `examples/bidi/static/`) lacked basic ARIA dialog semantics (`role="dialog"`, `aria-modal="true"`, `aria-labelledby`), keyboard focus trapping, Escape key dismissal, and focus restoration to the triggering button. Additionally, interactive buttons lacked `:focus-visible` styling, making keyboard navigation difficult to track for sighted keyboard users. Serving ES module scripts via `file://` triggers browser CORS restrictions during testing, requiring a local HTTP server for Playwright verification.

**Action:**
Ensure all custom modal implementations define WAI-ARIA dialog attributes, close on `Escape`, trap Tab focus within the active modal container, restore focus to the opener element upon dismissal, and provide distinct `:focus-visible` outlines on all interactive elements.
