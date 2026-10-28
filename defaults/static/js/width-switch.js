(function () {
  var STORAGE_KEY = "go-grip-width";
  // Filled from the data-width options in layout.html.
  var MODES = [];

  // Custom widths are any two-digit percentage.
  function isCustom(mode) {
    return /^[1-9][0-9]$/.test(mode);
  }

  function isValid(mode) {
    return MODES.indexOf(mode) !== -1 || isCustom(mode);
  }

  function label(mode) {
    return mode === "auto" ? "Auto" : mode + "%";
  }

  function getPreference() {
    try {
      var stored = localStorage.getItem(STORAGE_KEY);
      if (stored && isValid(stored)) return stored;
    } catch (e) {}
    return "auto";
  }

  function applyWidth(mode) {
    var container = document.querySelector("div.container");
    if (container) {
      // Both bounds, so widths below the default 896px max-width apply too.
      var width = mode === "auto" ? "" : mode + "%";
      container.style.minWidth = width;
      container.style.maxWidth = width;
    }

    var btn = document.getElementById("width-toggle");
    if (btn) {
      btn.title = "Width: " + label(mode);
    }

    var options = document.querySelectorAll("#width-panel [data-width]");
    for (var i = 0; i < options.length; i++) {
      options[i].classList.toggle("active", options[i].getAttribute("data-width") === mode);
    }

    var input = document.getElementById("width-custom-input");
    if (input) {
      var custom = MODES.indexOf(mode) === -1;
      // Don't rewrite the field while the user is typing in it.
      if (document.activeElement !== input) input.value = custom ? mode : "";
      input.parentNode.classList.toggle("active", custom);
    }

    try {
      localStorage.setItem(STORAGE_KEY, mode);
    } catch (e) {}
  }

  function setPanelOpen(open) {
    var btn = document.getElementById("width-toggle");
    var panel = document.getElementById("width-panel");
    if (!btn || !panel) return;
    panel.hidden = !open;
    btn.setAttribute("aria-expanded", open ? "true" : "false");
  }

  document.addEventListener("DOMContentLoaded", function () {
    var btn = document.getElementById("width-toggle");
    var panel = document.getElementById("width-panel");
    if (!btn || !panel) return;

    MODES = Array.prototype.map.call(panel.querySelectorAll("[data-width]"), function (option) {
      return option.getAttribute("data-width");
    });
    applyWidth(getPreference());

    btn.addEventListener("click", function () {
      setPanelOpen(panel.hidden);
    });

    panel.addEventListener("click", function (e) {
      var option = e.target.closest("[data-width]");
      if (!option) return;
      applyWidth(option.getAttribute("data-width"));
      setPanelOpen(false);
    });

    var input = document.getElementById("width-custom-input");
    input.addEventListener("input", function () {
      input.value = input.value.replace(/[^0-9]/g, "");
      if (isCustom(input.value)) applyWidth(input.value);
    });
    input.addEventListener("keydown", function (e) {
      if (e.key === "Enter" && isCustom(input.value)) setPanelOpen(false);
    });

    document.addEventListener("click", function (e) {
      if (!btn.contains(e.target) && !panel.contains(e.target)) {
        setPanelOpen(false);
      }
    });

    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape") setPanelOpen(false);
    });
  });
})();
