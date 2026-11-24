# Global Drag-and-Drop File Upload - Implementation Plan

**Task Slug:** `global-drag-drop`  
**Target:** Embedded Web Dashboard (`internal/web/static/index.html`)  
**Lead Agent:** `frontend-specialist` (with `project-planner` orchestration)  
**Applicable Skills:** `@antislop-ui`, `@antislop-human`, `@clean-code`, `@verify-changes`

---

## 1. Overview & Objective

Enable users to drag and drop files anywhere onto the web dashboard viewport to trigger seamless API-backed file uploads. Currently, drag-and-drop only triggers if the cursor targets the narrow `#dropZone` element; dropping anywhere else allows the browser's default behavior (navigating to/opening the file) to hijack the session.

---

## 2. Requirements & Scope

### Must-Have (P0)
- **Window-Level Drag Events (`dragenter`, `dragover`, `dragleave`, `drop`)**: Intercept and `preventDefault()` across `window` and `document` to prevent browser file-open navigation.
- **Nested Drag Event Counter**: Maintain an active drag target counter (`dragCounter`) to prevent flickering when the cursor moves over child DOM elements during a drag session.
- **Full-Screen Drop Overlay**: Dedicated overlay with high-contrast dashed boundary, accessible typography, and backdrop blur adhering to both light and dark themes.
- **Target Path Continuity**: Respect the existing "Upload path" input (`#customPath`) if populated by the user.
- **Toast Notifications**: Clear visual upload status and completion notifications.
- **Keyboard Escape Handling**: Allow pressing `Escape` to cancel and dismiss the active drag overlay if the user changes their mind.

### Nice-to-Have (P1 / P2)
- **Multi-File Batch Upload Queue**: Process multiple dropped files in a sequential pipeline rather than only `files[0]`.
- **Drag Rejection for Non-File Items**: Inspect `e.dataTransfer.types` so that dragging internal DOM text or links does not mistakenly trigger the upload overlay.

---

## 3. Technical Design

### A. Drag-and-Drop Event Lifecycle & Counter Pattern
In HTML5 Drag and Drop API, moving across nested elements triggers `dragleave` on the parent before `dragenter` on the child. Without a counter, the overlay flickers and disappears prematurely.

```javascript
let dragCounter = 0;

window.addEventListener('dragenter', (e) => {
  e.preventDefault();
  if (isFilesDrag(e)) {
    dragCounter++;
    showGlobalOverlay();
  }
});

window.addEventListener('dragleave', (e) => {
  e.preventDefault();
  dragCounter--;
  if (dragCounter <= 0) {
    dragCounter = 0;
    hideGlobalOverlay();
  }
});

window.addEventListener('dragover', (e) => {
  e.preventDefault();
  e.dataTransfer.dropEffect = 'copy';
});

window.addEventListener('drop', (e) => {
  e.preventDefault();
  dragCounter = 0;
  hideGlobalOverlay();
  if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files.length > 0) {
    handleFiles(e.dataTransfer.files);
  }
});
```

### B. Overlay DOM & Theming
```html
<div class="global-drop-overlay" id="globalDropOverlay" aria-hidden="true">
  <div class="global-drop-box">
    <svg class="drop-icon" ...></svg>
    <h3>Drop files anywhere to upload</h3>
    <p>Files will be uploaded to Simple Static CDN</p>
  </div>
</div>
```
- **Dark Mode**: High contrast border (`--border-focus: #3b82f6`), dark backdrop (`rgba(15, 17, 23, 0.85)`).
- **Light Mode**: High contrast border (`--border-focus: #2563eb`), crisp light backdrop (`rgba(248, 250, 252, 0.88)`).

---

## 4. Phase-by-Phase Task Breakdown

### Phase 1: CSS & Theming Setup
- [x] Add `.global-drop-overlay` and `.global-drop-box` CSS rules in `internal/web/static/index.html`.
- [x] Ensure token compliance with both `:root` (dark) and `[data-theme="light"]`.
- [x] Set `pointer-events: none` on inner elements to avoid child event bubbling anomalies.

### Phase 2: HTML Overlay Markup
- [x] Add `#globalDropOverlay` container before closing `</body>`.
- [x] Include accessible icon and clear instructional copy.

### Phase 3: Window Event Listener Wiring
- [x] Implement `isFilesDrag(e)` check inspecting `e.dataTransfer.types.includes('Files')`.
- [x] Bind `dragenter`, `dragover`, `dragleave`, `drop` to `window`.
- [x] Bind `keydown` listener for `Escape` to cancel and reset `dragCounter`.

### Phase 4: Batch Upload Queue Enhancement
- [x] Update upload handler (`handleBatchUpload`) to process multiple dropped files sequentially.
- [x] Provide progress toasts (`Uploading 1 of N: filename...`).
- [x] Refresh the file list (`loadFiles()`) once the batch completes.

### Phase 5: Verification & End-to-End Testing
- [x] Verify Go unit tests pass (`go test -v ./...`).
- [x] Verify live HTTP end-to-end suite (`verify.ps1`).
- [x] Theme and contrast compliance across dark and light modes.

---

## 5. Verification Checklist

| Check | Scenario | Expected Result |
|---|---|---|
| C1 | Drag file from Desktop over viewport | Fullscreen drop overlay smoothly appears |
| C2 | Move cursor between header, table, and dropzone | Overlay remains visible without flickering |
| C3 | Move cursor outside the browser window | Overlay smoothly disappears |
| C4 | Drop single file anywhere on page | File is uploaded via `/api/upload` and appears in table |
| C5 | Drop multiple files (e.g. 3 files) | All 3 files are uploaded with sequential toast updates |
| C6 | Drag selected text inside page | Overlay is NOT triggered |
| C7 | Press Escape while dragging | Overlay resets cleanly |
| C8 | Theme toggle between Dark and Light | Overlay colors adapt with full WCAG contrast |
