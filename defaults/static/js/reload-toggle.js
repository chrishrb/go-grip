// Per-page pause for live reload: while paused, live-reload.js records
// reload events here instead of reloading; resuming reloads the page only if
// something changed while paused.
(function () {
  var state = (window.gripReloadToggle = { paused: false, pending: false });

  // Inline SVGs instead of glyph characters, which render as color emoji
  // when the emoji font wins fallback.
  var ICON_PAUSE =
    '<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">' +
    '<rect x="3" y="2" width="3.5" height="12" rx="1"/><rect x="9.5" y="2" width="3.5" height="12" rx="1"/>' +
    "</svg>";
  var ICON_PLAY =
    '<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">' +
    '<path d="M3.5 2.5 L12.5 8 L3.5 13.5 Z"/>' +
    "</svg>";

  function updateButton() {
    var btn = document.getElementById("reload-toggle");
    if (!btn) return;
    var icon = btn.querySelector(".theme-toggle-icon");
    if (!icon) return;

    if (state.paused) {
      icon.innerHTML = ICON_PLAY;
      btn.title = "Resume auto-reload";
    } else {
      icon.innerHTML = ICON_PAUSE;
      btn.title = "Pause auto-reload";
    }
  }

  document.addEventListener("DOMContentLoaded", function () {
    var btn = document.getElementById("reload-toggle");
    if (!btn) return;
    updateButton();

    btn.addEventListener("click", function () {
      state.paused = !state.paused;
      if (!state.paused && state.pending) {
        window.location.reload();
        return;
      }
      updateButton();
    });
  });
})();
