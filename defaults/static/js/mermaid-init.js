(function(){
  var lightThemeVariables = {
    primaryColor: '#ECECFF',
    primaryTextColor: 'black',
    primaryBorderColor: 'hsl(259.6261682243, 59.7765363128%, 87.9019607843%)',
    lineColor: 'hsl(259.6261682243, 59.7765363128%, 87.9019607843%)'
  };

  var darkThemeVariables = {
    primaryColor: '#1f2020',
    primaryTextColor: 'lightgrey',
    primaryBorderColor: '#ccc',
    lineColor: '#ccc'
  };

  // SVG icons
  const ICON_ZOOM_IN = '<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-zoom-in" aria-hidden="true"><path d="M3.75 7.5a.75.75 0 0 1 .75-.75h2.25V4.5a.75.75 0 0 1 1.5 0v2.25h2.25a.75.75 0 0 1 0 1.5H8.25v2.25a.75.75 0 0 1-1.5 0V8.25H4.5a.75.75 0 0 1-.75-.75Z"></path><path d="M7.5 0a7.5 7.5 0 0 1 5.807 12.247l2.473 2.473a.749.749 0 1 1-1.06 1.06l-2.473-2.473A7.5 7.5 0 1 1 7.5 0Zm-6 7.5a6 6 0 1 0 12 0 6 6 0 0 0-12 0Z"></path></svg>';
  const ICON_ZOOM_OUT = '<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-zoom-out" aria-hidden="true"><path d="M4.5 6.75h6a.75.75 0 0 1 0 1.5h-6a.75.75 0 0 1 0-1.5Z"></path><path d="M0 7.5a7.5 7.5 0 1 1 13.307 4.747l2.473 2.473a.749.749 0 1 1-1.06 1.06l-2.473-2.473A7.5 7.5 0 0 1 0 7.5Zm7.5-6a6 6 0 1 0 0 12 6 6 0 0 0 0-12Z"></path></svg>';
  const ICON_RESET = '<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-sync" aria-hidden="true"><path d="M1.705 8.005a.75.75 0 0 1 .834.656 5.5 5.5 0 0 0 9.592 2.97l-1.204-1.204a.25.25 0 0 1 .177-.427h3.646a.25.25 0 0 1 .25.25v3.646a.25.25 0 0 1-.427.177l-1.38-1.38A7.002 7.002 0 0 1 1.05 8.84a.75.75 0 0 1 .656-.834ZM8 2.5a5.487 5.487 0 0 0-4.131 1.869l1.204 1.204A.25.25 0 0 1 4.896 6H1.25A.25.25 0 0 1 1 5.75V2.104a.25.25 0 0 1 .427-.177l1.38 1.38A7.002 7.002 0 0 1 14.95 7.16a.75.75 0 0 1-1.49.178A5.5 5.5 0 0 0 8 2.5Z"></path></svg>';
  const ICON_PAN_UP = '<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-chevron-up" aria-hidden="true"><path d="M3.22 10.53a.749.749 0 0 1 0-1.06l4.25-4.25a.749.749 0 0 1 1.06 0l4.25 4.25a.749.749 0 1 1-1.06 1.06L8 6.811 4.28 10.53a.749.749 0 0 1-1.06 0Z"></path></svg>';
  const ICON_PAN_DOWN = '<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-chevron-down" aria-hidden="true"><path d="M12.78 5.22a.749.749 0 0 1 0 1.06l-4.25 4.25a.749.749 0 0 1-1.06 0L3.22 6.28a.749.749 0 1 1 1.06-1.06L8 8.939l3.72-3.719a.749.749 0 0 1 1.06 0Z"></path></svg>';
  const ICON_PAN_LEFT = '<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-chevron-left" aria-hidden="true"><path d="M9.78 12.78a.75.75 0 0 1-1.06 0L4.47 8.53a.75.75 0 0 1 0-1.06l4.25-4.25a.751.751 0 0 1 1.042.018.751.751 0 0 1 .018 1.042L6.06 8l3.72 3.72a.75.75 0 0 1 0 1.06Z"></path></svg>'
  const ICON_PAN_RIGHT = '<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-chevron-right" aria-hidden="true"><path d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06Z"></path></svg>'

  function isDark() {
    var theme = document.body.getAttribute('data-theme');
    if (theme === 'dark') return true;
    if (theme === 'light') return false;
    return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
  }

  function escapeHTML(str) {
    return String(str).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
  }

  function getCode(node) {
    return node.dataset.code || node.textContent || '';
  }

  function createBtn(html, className, title, fn) {
    var b = document.createElement('button');
    b.className = `mermaid-toolbar-btn ${className}`;
    b.innerHTML = html;
    b.title = title;
    b.type = 'button';
    b.addEventListener('click', function(e) { e.preventDefault(); e.stopPropagation(); fn(); });
    return b;
  }

  function attachZoomPan(viewport, svg) {
    svg.style.transformOrigin = '0 0';

    var scale = 1, panX = 0, panY = 0;
    var ZOOM_STEP = 0.2;
    var PAN_STEP = 50;
    var MIN_ZOOM = 0.1;
    var MAX_ZOOM = 10;

    function apply() {
      svg.style.transform = 'translate(' + panX + 'px,' + panY + 'px) scale(' + scale + ')';
    }

    function zoomAt(newScale, cx, cy) {
      newScale = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, newScale));
      var ratio = newScale / scale;
      panX = cx - (cx - panX) * ratio;
      panY = cy - (cy - panY) * ratio;
      scale = newScale;
      apply();
    }

    function zoomCenter(delta) {
      zoomAt(scale + delta, viewport.clientWidth / 2, viewport.clientHeight / 2);
    }

    var buttons = [];
    buttons.push(createBtn(ICON_ZOOM_IN, 'zoom-in', 'Zoom in', function() { zoomCenter(ZOOM_STEP); }));
    buttons.push(createBtn(ICON_ZOOM_OUT, 'zoom-out', 'Zoom out', function() { zoomCenter(-ZOOM_STEP); }));
    buttons.push(createBtn(ICON_RESET, 'reset', 'Reset view', function() { scale = 1; panX = 0; panY = 0; apply(); }));
    buttons.push(createBtn(ICON_PAN_UP, 'up', 'Pan up', function() { panY += PAN_STEP; apply(); }));
    buttons.push(createBtn(ICON_PAN_DOWN, 'down', 'Pan down', function() { panY -= PAN_STEP; apply(); }));
    buttons.push(createBtn(ICON_PAN_LEFT, 'left', 'Pan left', function() { panX += PAN_STEP; apply(); }));
    buttons.push(createBtn(ICON_PAN_RIGHT, 'right', 'Pan right', function() { panX -= PAN_STEP; apply(); }));

    return buttons;
  }

  function wrapWithControls(node) {
    var svg = node.querySelector('svg');
    if (!svg) return;

    // Use code for copy button
    // var code = node.dataset.code || '';

    var viewport = document.createElement('div');
    viewport.className = 'mermaid-viewport';

    svg.parentNode.removeChild(svg);
    viewport.appendChild(svg);

    var toolbar = document.createElement('div');
    toolbar.className = 'mermaid-viewer-control-panel';

    var zoomBtns = attachZoomPan(viewport, svg);
    for (var i = 0; i < zoomBtns.length; i++) toolbar.appendChild(zoomBtns[i]);

    node.appendChild(toolbar);
    node.appendChild(viewport);
  }

  async function renderAll() {
    if (!window.mermaid) return;

    var themeVars = isDark() ? darkThemeVariables : lightThemeVariables;
    mermaid.initialize({
      startOnLoad: false,
      theme: 'base',
      themeVariables: themeVars,
      securityLevel: 'loose',
      logLevel: 'error'
    });

    var nodes = document.querySelectorAll('.mermaid');
    for (var i = 0; i < nodes.length; i++) {
      var node = nodes[i];
      var code = getCode(node).trim();
      // Preserve source code for re-renders
      if (!node.dataset.code) node.dataset.code = code;

      try {
        var id = 'mermaid-svg-' + i + '-' + Date.now();
        node.classList.remove('rendered');
        var result = await mermaid.render(id, code);
        node.innerHTML = result.svg;
        wrapWithControls(node);
        node.classList.remove('mermaid-error');
        node.classList.add('rendered');
      } catch(err) {
        console.error('Mermaid render error:', err);
        node.classList.add('mermaid-error');
        node.innerHTML = '<pre class="mermaid-error">Mermaid error:\n' + escapeHTML(err.message || String(err)) + '</pre>';
      }
    }
  }

  document.addEventListener('DOMContentLoaded', function() {
    renderAll();

    if (window.matchMedia) {
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', renderAll);
    }
    document.body.addEventListener('themechange', renderAll);
  });
})();
