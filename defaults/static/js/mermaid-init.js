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

  function wrapWithControls(node) {
    var svg = node.querySelector('svg');
    if (!svg) return;

    var viewport = document.createElement('div');
    viewport.className = 'mermaid-viewport';

    svg.parentNode.removeChild(svg);
    svg.style.transformOrigin = '0 0';
    viewport.appendChild(svg);

    var toolbar = document.createElement('div');
    toolbar.className = 'mermaid-toolbar';

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

    function btn(html, title, fn) {
      var b = document.createElement('button');
      b.className = 'mermaid-toolbar-btn';
      b.innerHTML = html;
      b.title = title;
      b.type = 'button';
      b.addEventListener('click', function(e) { e.preventDefault(); e.stopPropagation(); fn(); });
      toolbar.appendChild(b);
    }

    btn('&#x2212;', 'Zoom out', function() { zoomCenter(-ZOOM_STEP); });
    btn('&#x2b;', 'Zoom in', function() { zoomCenter(ZOOM_STEP); });
    btn('&#x2190;', 'Pan left', function() { panX += PAN_STEP; apply(); });
    btn('&#x2192;', 'Pan right', function() { panX -= PAN_STEP; apply(); });
    btn('&#x2191;', 'Pan up', function() { panY += PAN_STEP; apply(); });
    btn('&#x2193;', 'Pan down', function() { panY -= PAN_STEP; apply(); });
    btn('&#x21bb;', 'Reset', function() { scale = 1; panX = 0; panY = 0; apply(); });

    // Wheel zoom at cursor position
    viewport.addEventListener('wheel', function(e) {
      e.preventDefault();
      var rect = viewport.getBoundingClientRect();
      var delta = e.deltaY < 0 ? ZOOM_STEP : -ZOOM_STEP;
      zoomAt(scale + delta, e.clientX - rect.left, e.clientY - rect.top);
    }, { passive: false });

    // Drag to pan
    var dragging = false, lastX = 0, lastY = 0;
    viewport.addEventListener('pointerdown', function(e) {
      if (e.button !== 0) return;
      dragging = true;
      lastX = e.clientX;
      lastY = e.clientY;
      viewport.setPointerCapture(e.pointerId);
      viewport.classList.add('dragging');
    });
    viewport.addEventListener('pointermove', function(e) {
      if (!dragging) return;
      panX += e.clientX - lastX;
      panY += e.clientY - lastY;
      lastX = e.clientX;
      lastY = e.clientY;
      apply();
    });
    viewport.addEventListener('pointerup', function(e) {
      if (!dragging) return;
      dragging = false;
      viewport.releasePointerCapture(e.pointerId);
      viewport.classList.remove('dragging');
    });

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
