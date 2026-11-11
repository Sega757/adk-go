## 2025-05-18 - Accessible Modal Dialogs and Focus Management in Vanilla JavaScript Demos

**Learning:** In lightweight static HTML/JS web demos without UI framework components, custom modal overlays often omit `role="dialog"`, `aria-modal="true"`, `aria-labelledby`, and keyboard focus trapping. This traps screen reader users in the underlying background DOM and leaves keyboard-only users without `Escape` dismissal or focus restoration.
**Action:** When implementing or enhancing modal dialogues in vanilla frontend assets, always add ARIA dialog semantics, bind an `Escape` key handler, implement cyclic `Tab`/`Shift+Tab` focus wrapping within active modal buttons, and save/restore `document.activeElement`.
