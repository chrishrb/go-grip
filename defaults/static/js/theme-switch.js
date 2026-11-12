(function () {
  var STORAGE_KEY = "go-grip-theme";
  var MODES = ["light", "dark"];
  var CSS_SPACE = "(?:\\s|/\\*(?:[^*]|\\*(?!/))*\\*/)*";
  var SCHEME_CONDITION = new RegExp(
    "(\\(" + CSS_SPACE + "prefers-color-scheme" + CSS_SPACE + ":" + CSS_SPACE + ")(light|dark)(" + CSS_SPACE + "\\))",
    "gi"
  );
  var originalMedia = new WeakMap();

  function browserPrefers() {
    if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
      return "dark";
    }
    return "light";
  }

  function getPreference() {
    try {
      var stored = localStorage.getItem(STORAGE_KEY);
      if (stored && MODES.indexOf(stored) !== -1) return stored;
    } catch (e) {}
    return browserPrefers();
  }

  function swapScheme(match, before, scheme, after) {
    return before + (scheme.toLowerCase() === "dark" ? "light" : "dark") + after;
  }

  function syncPictureSources(mode) {
    var swap = mode !== browserPrefers();
    var sources = document.querySelectorAll("picture > source[media]");

    for (var i = 0; i < sources.length; i++) {
      var source = sources[i];
      if (!originalMedia.has(source)) {
        if (source.media.search(SCHEME_CONDITION) === -1) continue;
        originalMedia.set(source, source.media);
      }

      var original = originalMedia.get(source);
      var media = swap ? original.replace(SCHEME_CONDITION, swapScheme) : original;
      if (source.media !== media) source.media = media;
    }
  }

  function applyTheme(mode) {
    var lightCSS = document.getElementById("theme-light");
    var darkCSS = document.getElementById("theme-dark");
    var lightHL = document.getElementById("highlight-light");
    var darkHL = document.getElementById("highlight-dark");

    if (mode === "dark") {
      if (lightCSS) lightCSS.media = "not all";
      if (darkCSS) darkCSS.media = "all";
      if (lightHL) lightHL.media = "not all";
      if (darkHL) darkHL.media = "all";
      document.body.setAttribute("data-theme", "dark");
    } else {
      if (lightCSS) lightCSS.media = "all";
      if (darkCSS) darkCSS.media = "not all";
      if (lightHL) lightHL.media = "all";
      if (darkHL) darkHL.media = "not all";
      document.body.setAttribute("data-theme", "light");
    }

    syncPictureSources(mode);
    updateIcon(mode);

    try {
      localStorage.setItem(STORAGE_KEY, mode);
    } catch (e) {}

    document.body.dispatchEvent(
      new CustomEvent("themechange", { detail: { mode: mode } })
    );
  }

  function updateIcon(mode) {
    var btn = document.getElementById("theme-toggle");
    if (!btn) return;
    var icon = btn.querySelector(".theme-toggle-icon");
    if (!icon) return;

    if (mode === "dark") {
      icon.textContent = "\u263E";
      btn.title = "Theme: Dark";
    } else {
      icon.textContent = "\u2600";
      btn.title = "Theme: Light";
    }
  }

  function toggle() {
    var current = getPreference();
    applyTheme(current === "light" ? "dark" : "light");
  }

  document.addEventListener("DOMContentLoaded", function () {
    applyTheme(getPreference());

    var btn = document.getElementById("theme-toggle");
    if (btn) {
      btn.addEventListener("click", toggle);
    }

    if (window.matchMedia) {
      window
        .matchMedia("(prefers-color-scheme: dark)")
        .addEventListener("change", function () {
          syncPictureSources(document.body.getAttribute("data-theme"));
        });
    }
  });
})();
