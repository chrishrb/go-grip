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
  var ICON_COPY = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="5.5" y="1.5" width="9" height="9" rx="1.5"/><path d="M1.5 5.5v8c0 .83.67 1.5 1.5 1.5h8"/></svg>';
  var ICON_CHECK = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#3fb950" stroke-width="2"><path d="M3 8l3.5 3.5L13 5"/></svg>';
  var ICON_EXPAND = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9.5 1.5h5v5M6.5 14.5h-5v-5M14 2l-5 5M2 14l5-5"/></svg>';

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

  function createBtn(html, title, fn) {
    var b = document.createElement('button');
    b.className = 'mermaid-toolbar-btn';
    b.innerHTML = html;
    b.title = title;
    b.type = 'button';
    b.addEventListener('click', function(e) { e.preventDefault(); e.stopPropagation(); fn(); });
    return b;
  }

  function createSep() {
    var s = document.createElement('span');
    s.className = 'mermaid-toolbar-sep';
    return s;
  }

  function createCopyBtn(code) {
    var btn = createBtn(ICON_COPY, 'Copy diagram code', function() {
      if (!navigator.clipboard) return;
      navigator.clipboard.writeText(code).then(function() {
        btn.innerHTML = ICON_CHECK;
        setTimeout(function() { btn.innerHTML = ICON_COPY; }, 2000);
      });
    });
    return btn;
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
    buttons.push(createBtn('&#x2212;', 'Zoom out', function() { zoomCenter(-ZOOM_STEP); }));
    buttons.push(createBtn('&#x2b;', 'Zoom in', function() { zoomCenter(ZOOM_STEP); }));
    buttons.push(createBtn('&#x2190;', 'Pan left', function() { panX += PAN_STEP; apply(); }));
    buttons.push(createBtn('&#x2192;', 'Pan right', function() { panX -= PAN_STEP; apply(); }));
    buttons.push(createBtn('&#x2191;', 'Pan up', function() { panY += PAN_STEP; apply(); }));
    buttons.push(createBtn('&#x2193;', 'Pan down', function() { panY -= PAN_STEP; apply(); }));
    buttons.push(createBtn('&#x21bb;', 'Reset', function() { scale = 1; panX = 0; panY = 0; apply(); }));

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

    return buttons;
  }

  function openModal(svgElement, code) {
    var overlay = document.createElement('div');
    overlay.className = 'mermaid-modal-overlay';

    var modal = document.createElement('div');
    modal.className = 'mermaid-modal';

    var header = document.createElement('div');
    header.className = 'mermaid-modal-header';

    var toolbar = document.createElement('div');
    toolbar.className = 'mermaid-toolbar';

    var viewport = document.createElement('div');
    viewport.className = 'mermaid-modal-viewport';

    var svg = svgElement.cloneNode(true);
    svg.style.transform = '';
    svg.style.transformOrigin = '';
    viewport.appendChild(svg);

    var zoomBtns = attachZoomPan(viewport, svg);
    for (var i = 0; i < zoomBtns.length; i++) toolbar.appendChild(zoomBtns[i]);

    toolbar.appendChild(createSep());
    toolbar.appendChild(createCopyBtn(code));

    var closeBtn = document.createElement('button');
    closeBtn.className = 'mermaid-modal-close';
    closeBtn.innerHTML = '&times;';
    closeBtn.title = 'Close (Esc)';
    closeBtn.type = 'button';

    header.appendChild(toolbar);
    header.appendChild(closeBtn);

    modal.appendChild(header);
    modal.appendChild(viewport);
    overlay.appendChild(modal);
    document.body.appendChild(overlay);
    document.body.style.overflow = 'hidden';

    function close() {
      document.body.removeChild(overlay);
      document.body.style.overflow = '';
      document.removeEventListener('keydown', keyHandler);
    }

    function keyHandler(e) {
      if (e.key === 'Escape') close();
    }

    closeBtn.addEventListener('click', close);
    overlay.addEventListener('click', function(e) { if (e.target === overlay) close(); });
    document.addEventListener('keydown', keyHandler);

    requestAnimationFrame(function() { closeBtn.focus(); });
  }

  function wrapWithControls(node) {
    var svg = node.querySelector('svg');
    if (!svg) return;

    var code = node.dataset.code || '';

    var viewport = document.createElement('div');
    viewport.className = 'mermaid-viewport';

    svg.parentNode.removeChild(svg);
    viewport.appendChild(svg);

    var toolbar = document.createElement('div');
    toolbar.className = 'mermaid-toolbar';

    var zoomBtns = attachZoomPan(viewport, svg);
    for (var i = 0; i < zoomBtns.length; i++) toolbar.appendChild(zoomBtns[i]);

    toolbar.appendChild(createSep());
    toolbar.appendChild(createCopyBtn(code));
    toolbar.appendChild(createBtn(ICON_EXPAND, 'Open in fullscreen', function() {
      openModal(svg, code);
    }));

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
