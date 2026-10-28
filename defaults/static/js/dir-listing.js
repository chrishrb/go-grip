(function () {
  function rowLinks() {
    return Array.prototype.slice.call(
      document.querySelectorAll(".grip-listing-row .grip-listing-name a")
    );
  }

  function focusLink(link) {
    if (!link) return;
    link.focus({ preventScroll: true });
    var row = link.closest(".grip-listing-row") || link;
    row.scrollIntoView({ block: "nearest" });
  }

  function upLink() {
    return document.querySelector(".grip-listing-up a");
  }

  function initialFocus(links) {
    var params;
    try {
      params = new URLSearchParams(window.location.search);
    } catch (e) {
      params = null;
    }
    var from = params ? params.get("from") : null;
    var target = null;

    if (from !== null) {
      for (var i = 0; i < links.length; i++) {
        if (links[i].textContent === from) {
          target = links[i];
          break;
        }
      }
      params.delete("from");
      var query = params.toString();
      var url =
        window.location.pathname + (query ? "?" + query : "") + window.location.hash;
      try {
        window.history.replaceState(window.history.state, "", url);
      } catch (e) {}
    }

    focusLink(target || links[0]);
  }

  function isFormField(el) {
    if (!el || !el.tagName) return false;
    var tag = el.tagName.toLowerCase();
    return tag === "input" || tag === "textarea" || tag === "select" || el.isContentEditable;
  }

  function onKeyDown(event) {
    if (event.ctrlKey || event.altKey || event.metaKey) return;
    if (isFormField(event.target)) return;

    var links = rowLinks();
    if (links.length === 0) return;
    var current = links.indexOf(document.activeElement);

    switch (event.key) {
      case "ArrowDown":
      case "j":
        focusLink(links[current < 0 ? 0 : Math.min(current + 1, links.length - 1)]);
        break;
      case "ArrowUp":
      case "k":
        focusLink(links[current < 0 ? 0 : Math.max(current - 1, 0)]);
        break;
      case "Home":
        focusLink(links[0]);
        break;
      case "End":
        focusLink(links[links.length - 1]);
        break;
      case "Backspace":
      case "ArrowLeft":
        var up = upLink();
        if (up) window.location.href = up.href;
        break;
      default:
        return;
    }
    event.preventDefault();
  }

  document.addEventListener("DOMContentLoaded", function () {
    var links = rowLinks();
    if (links.length === 0) return;
    initialFocus(links);
    document.addEventListener("keydown", onKeyDown);
  });
})();
