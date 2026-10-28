// Live reload: go-grip streams a "reload" event when something this page
// shows has changed (the Markdown file, an image it references, or for a
// folder page, the folder's entries).
(function () {
  if (!window.EventSource) return;

  var pagePath = location.pathname;
  try {
    pagePath = decodeURIComponent(pagePath);
  } catch (e) {}
  var url = "/__grip/events?path=" + encodeURIComponent(pagePath);

  var source = null;

  function connect() {
    var es = new EventSource(url);
    var lostConnection = false;

    es.addEventListener("reload", function () {
      es.close();
      location.reload();
    });

    es.onerror = function () {
      lostConnection = true;
    };

    // The connection came back after an error: the server probably
    // restarted, so the page may be stale.
    es.onopen = function () {
      if (lostConnection) {
        es.close();
        location.reload();
      }
    };

    source = es;
  }

  // Leaving the page. Chrome's back/forward cache freezes the page instead of
  // unloading it, and would keep this stream's connection open. A few frozen
  // pages use up the browser's per-host connection limit and stall the next
  // navigation, so release the connection. Closing an EventSource keeps the
  // page eligible for the cache; an unload handler would not.
  window.addEventListener("pagehide", function () {
    if (source) {
      source.close();
      source = null;
    }
  });

  // Back from the cache. Files may have changed while the page was frozen
  // and the server cannot tell which ones, so load the page fresh. On
  // localhost that costs one render, and the new page opens its own stream.
  window.addEventListener("pageshow", function (event) {
    if (event.persisted) {
      location.reload();
    }
  });

  connect();
})();
