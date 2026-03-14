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

  var ZOOM_STEP = 0.2;
  var PAN_STEP = 50;
  var MIN_ZOOM = 0.1;
  var MAX_ZOOM = 10;

  function initZoomPanState(viewport, svg) {
    svg.style.transformOrigin = '0 0';

    var state = { scale: 1, panX: 0, panY: 0 };

    function apply() {
      svg.style.transform = 'translate(' + state.panX + 'px,' + state.panY + 'px) scale(' + state.scale + ')';
    }

    function zoomAt(newScale, cx, cy) {
      newScale = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, newScale));
      var ratio = newScale / state.scale;
      state.panX = cx - (cx - state.panX) * ratio;
      state.panY = cy - (cy - state.panY) * ratio;
      state.scale = newScale;
      apply();
    }

    function zoomCenter(delta) {
      zoomAt(state.scale + delta, viewport.clientWidth / 2, viewport.clientHeight / 2);
    }

    // Wheel zoom at cursor position
    viewport.addEventListener('wheel', function(e) {
      e.preventDefault();
      var rect = viewport.getBoundingClientRect();
      var delta = e.deltaY < 0 ? ZOOM_STEP : -ZOOM_STEP;
      zoomAt(state.scale + delta, e.clientX - rect.left, e.clientY - rect.top);
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
      state.panX += e.clientX - lastX;
      state.panY += e.clientY - lastY;
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

    return {
      zoomCenter: zoomCenter,
      apply: apply,
      state: state
    };
  }

  // Create toolbar buttons dynamically (used for inline diagrams)
  function attachZoomPan(viewport, svg) {
    var ctrl = initZoomPanState(viewport, svg);

    var buttons = [];
    buttons.push(createBtn('&#x2212;', 'Zoom out', function() { ctrl.zoomCenter(-ZOOM_STEP); }));
    buttons.push(createBtn('&#x2b;', 'Zoom in', function() { ctrl.zoomCenter(ZOOM_STEP); }));
    buttons.push(createBtn('&#x2190;', 'Pan left', function() { ctrl.state.panX += PAN_STEP; ctrl.apply(); }));
    buttons.push(createBtn('&#x2192;', 'Pan right', function() { ctrl.state.panX -= PAN_STEP; ctrl.apply(); }));
    buttons.push(createBtn('&#x2191;', 'Pan up', function() { ctrl.state.panY += PAN_STEP; ctrl.apply(); }));
    buttons.push(createBtn('&#x2193;', 'Pan down', function() { ctrl.state.panY -= PAN_STEP; ctrl.apply(); }));
    buttons.push(createBtn('&#x21bb;', 'Reset', function() { ctrl.state.scale = 1; ctrl.state.panX = 0; ctrl.state.panY = 0; ctrl.apply(); }));

    return buttons;
  }

  // Wire up data-action buttons from a template (used for modal)
  function attachZoomPanFromTemplate(container, viewport, svg) {
    var ctrl = initZoomPanState(viewport, svg);

    var actions = {
      'zoom-out': function() { ctrl.zoomCenter(-ZOOM_STEP); },
      'zoom-in': function() { ctrl.zoomCenter(ZOOM_STEP); },
      'pan-left': function() { ctrl.state.panX += PAN_STEP; ctrl.apply(); },
      'pan-right': function() { ctrl.state.panX -= PAN_STEP; ctrl.apply(); },
      'pan-up': function() { ctrl.state.panY += PAN_STEP; ctrl.apply(); },
      'pan-down': function() { ctrl.state.panY -= PAN_STEP; ctrl.apply(); },
      'reset': function() { ctrl.state.scale = 1; ctrl.state.panX = 0; ctrl.state.panY = 0; ctrl.apply(); }
    };

    var btns = container.querySelectorAll('[data-action]');
    for (var i = 0; i < btns.length; i++) {
      var action = btns[i].getAttribute('data-action');
      if (actions[action]) {
        btns[i].addEventListener('click', (function(fn) {
          return function(e) { e.preventDefault(); e.stopPropagation(); fn(); };
        })(actions[action]));
      }
    }
  }

  function openModal(svgElement, code) {
    var tpl = document.getElementById('mermaid-modal-template');
    if (!tpl) return;

    var fragment = tpl.content.cloneNode(true);
    var overlay = fragment.querySelector('.mermaid-modal-overlay');
    var viewport = fragment.querySelector('.mermaid-modal-viewport');
    var closeBtn = fragment.querySelector('.mermaid-modal-close');

    // Insert cloned SVG into viewport
    var svg = svgElement.cloneNode(true);
    svg.style.transform = '';
    svg.style.transformOrigin = '';
    viewport.appendChild(svg);

    // Wire up zoom/pan buttons from the template
    attachZoomPanFromTemplate(overlay, viewport, svg);

    // Wire up copy button
    var copyBtn = overlay.querySelector('[data-action="copy"]');
    if (copyBtn) {
      copyBtn.addEventListener('click', function(e) {
        e.preventDefault();
        e.stopPropagation();
        if (!navigator.clipboard) return;
        navigator.clipboard.writeText(code).then(function() {
          copyBtn.innerHTML = ICON_CHECK;
          setTimeout(function() { copyBtn.innerHTML = ICON_COPY; }, 2000);
        });
      });
    }

    document.body.appendChild(fragment);
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
