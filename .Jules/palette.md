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

## 2026-09-19 - Interactive Chat Image Lightbox Modal & Focus Restoration
**Learning:** In chat interfaces supporting camera snapshots or file attachments, static image thumbnails (max 300px height) prevent users from inspecting fine details, reading captured text, or verifying uploaded files. Making image thumbnails keyboard-focusable (`tabindex="0"`, `role="button"`) and triggering an accessible modal lightbox (`#imageModal`) with Escape key listeners and focus restoration allows mouse and keyboard users to inspect high-resolution images effortlessly without leaving the conversation view.
**Action:** Always make inline media attachments interactive by binding `click` and `Enter`/`Space` key listeners, setting descriptive `aria-label` and `title` tooltips, and managing focus restoration when opening modal lightboxes.

## 2026-09-20 - Dynamic Unread Message Counter for Floating Scroll-to-Bottom Controls
**Learning:** For real-time streaming chat interfaces where auto-scroll is paused while a user reviews past message history, displaying a static jump button without message counts leaves users unaware of incoming stream volume. Dynamically tracking unread messages while `isNearBottom` evaluates to false and updating the scroll button's text (`↓ 3 new messages`), `aria-label`, and `title` tooltip provides immediate visual and assistive feedback on new activity without disrupting reading position.
**Action:** When implementing floating jump-to-bottom buttons for streaming feeds, track unread arrivals while scrolled away from bottom and update button text and ARIA labels with the unread count.

## 2026-10-02 - Nested Interactive Controls inside Expandable Buttons & Keyboard Target Validation
**Learning:** When embedding nested interactive controls (such as a "Copy JSON" button) inside an expandable card or container element assigned `role="button"` with keyboard navigation, clicking or pressing `Enter`/`Space` on the child button can trigger event bubbling that toggles the parent collapse/expand state unexpectedly. Calling `e.stopPropagation()` on child button click handlers and validating `e.target === entry` in the parent `keydown` listener prevents keyboard activations and mouse clicks on child actions from accidentally collapsing or expanding the parent container.
**Action:** When embedding child action buttons inside interactive `role="button"` parent containers, always call `e.stopPropagation()` on child click handlers and verify `e.target === container` in the parent `keydown` listener.

## 2026-10-03 - Focus Preservation on Transient or Disabling UI Controls
**Learning:** When interactive elements (such as a floating "scroll-to-bottom" pill or a "clear console" action button) are hidden (`display: none`) or disabled (`disabled = true`) directly upon user activation, leaving standard browser focus on the newly hidden or disabled element traps keyboard navigation or causes focus to drop unexpectedly. Shifting focus explicitly to the active text input (`#message`) or an adjacent enabled control (`#showAudioEvents`) ensures seamless keyboard flow and uninterrupted user interaction.
**Action:** Whenever a button or control is hidden or disabled as a result of its own click/keydown action, explicitly transfer focus to a logical enabled input or adjacent interactive element.

## 2026-10-04 - Clipboard Image Pasting Support in Web Chat Input Controls
**Learning:** In multimodal AI chat applications, users frequently copy screenshots or web images to their system clipboard. Forcing users to save the image to disk and navigate a file selection dialog or drag-and-drop zone creates unnecessary friction. Intercepting standard `paste` events on text input fields, inspecting `e.clipboardData.items` for `image/*` MIME types, and converting pasted images into interactive chat bubbles with accessible alt text ("Pasted image from clipboard") creates a smooth, intuitive keyboard-friendly media submission workflow.
**Action:** When designing text input fields in multimodal chat interfaces, attach a `paste` event listener to check `clipboardData.items` for image payloads, prevent default text paste behavior, render a preview thumbnail, and send the base64 image data to the server.
