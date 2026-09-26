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

## 2026-09-04 - Dynamic Input Control Feedback and Title Tooltips for Disconnected States
**Learning:** Disabling web form controls during disconnected or connecting network states without informative tooltip hints (`title`) or placeholder updates leaves screen reader and mouse users confused about why controls are un-interactive. Dynamically setting title tooltips (e.g., `"Connect to server to enable control"`) and input placeholder text (`"Connecting to server..."`) when disabled—and clearing them upon reconnection—provides clear context and reduces user friction.
**Action:** When toggling input element `disabled` states based on WebSocket connection status, update input placeholders and add/remove descriptive `title` tooltips dynamically.

## 2026-09-05 - Dynamic Form Button States & Contextual Tooltip Hints for Empty States
**Learning:** Leaving action buttons (like "Send" or "Clear Console") enabled when their underlying inputs or target containers are empty leads to silent no-op clicks and user confusion. Dynamically updating `disabled` attributes alongside contextual `title` tooltips (e.g. `"Type a message to enable send"` or `"Console is empty"`) as users type, submit, or clear data provides instant visual and screen-reader feedback, preventing accidental empty submissions.
**Action:** Always synchronize action button `disabled` states and descriptive `title` tooltips with input values and container content lengths.

## 2026-09-14 - Agent Chat Bubble Copy-to-Clipboard Action with Focus-Within and Transient ARIA Feedback
**Learning:** In streaming chat interfaces, copying agent text responses manually with mouse text selection can be cumbersome on touch devices or custom styled bubbles. Attaching a dedicated copy button inside agent bubbles that is revealed via CSS `:hover` or `:focus-within` allows keyboard navigators and mouse users to easily copy text. Dynamically toggling the button's `aria-label` and `title` to `"Copied to clipboard"` alongside visual confirmation (`✓`) for 2 seconds gives immediate multi-sensory feedback without disrupting chat layout.
**Action:** When adding inline utility actions to chat bubbles or card elements, use absolute positioning paired with parent `:focus-within` and `:focus-visible` rules for keyboard access, and temporarily swap `aria-label` and `title` attributes upon async completion.

## 2026-09-16 - Smart Auto-Scrolling & Floating Jump-to-Bottom Controls during Live Response Streaming
**Learning:** In real-time streaming chat interfaces, forcing automatic scroll-to-bottom on every incoming chunk disrupts users who scroll up to review previous conversation history. Pausing auto-scroll when `isNearBottom` evaluates to false (e.g. `diff > 100px`) while revealing a floating action pill button (`↓ New messages`) with explicit `aria-label` and `title` tooltips allows users to read earlier messages uninterrupted and jump back to the bottom in one click.
**Action:** Always check whether a scrolling container is near the bottom before auto-scrolling during streaming events, force-scroll on user actions, and pair auto-scroll pauses with a floating scroll-to-bottom button.

## 2026-09-17 - Inline Input Clearing via Escape Key & Keyboard Shortcut Tooltips in Web Chat Forms
**Learning:** In interactive web chat interfaces, users drafting messages often press the Escape key when they want to discard their current text. Without an explicit `keydown` listener, pressing Escape in a text input is a no-op, forcing tedious manual backspacing. Adding an `Escape` key event handler that clears the input and updates submit button states, paired with informative `title` tooltips (`"Type a message and press Enter to send (Esc to clear)"`), provides intuitive keyboard ergonomics.
**Action:** When creating text input fields in web chat or form interfaces, attach an `Escape` key event listener to quickly clear non-empty draft values, and document keyboard interaction shortcuts in element `title` tooltips.

## 2026-09-18 - Drag-and-Drop Image File Uploading with Visual Drop Zone Highlighting
**Learning:** For web chat interfaces that support file attachments, requiring users to exclusively click a file picker button increases friction. Attaching `dragenter`, `dragover`, `dragleave`, and `drop` event listeners to the input container with a drag counter (to prevent hover flickering across child elements) and applying an accented dashed border (`.drag-over`) provides instant visual cues and a seamless drop target for uploading image files.
**Action:** When adding drag-and-drop file upload capabilities, apply drag-and-drop event listeners to the container element, track hover state with a drag counter, toggle a CSS `.drag-over` focus class with clear border styling, and sanitize/validate dropped files before processing.

## 2026-09-19 - File Attachment MIME Type Validation & System Feedback Guidance
**Learning:** In web chat interfaces supporting image attachments, allowing users to select unsupported file types via standard file picker dialogs without client-side MIME type validation leads to broken visual thumbnails and invalid payload transmissions. Validating `file.type.startsWith('image/')` early in `fileInput` change listeners, displaying clear system message feedback (`Please select an image file (JPEG or PNG)`), and resetting input selection prevents broken visual layouts and invalid payload transmissions.
**Action:** Always perform early MIME type validation on file input change events, surface user-friendly system feedback messages when invalid files are selected, and reset input state to maintain interface cleanliness.
