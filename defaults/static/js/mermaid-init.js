(function(){
  function computeTheme(){
    try {
      var bodyTheme = document.body.getAttribute('data-theme') || 'auto';
      if (bodyTheme === 'dark') return 'dark';
      if (bodyTheme === 'light') return 'default';
      // auto
      var prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
      return prefersDark ? 'dark' : 'default';
    } catch(e){
      return 'default';
    }
  }

  function escapeHTML(str){
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function getCode(node){
    return (node.dataset && node.dataset.code) ? node.dataset.code : (node.textContent || '');
  }
  function setCode(node, code){
    if (node.dataset) node.dataset.code = code;
  }

  function renderWithAPI(id, code, node){
    return new Promise(function(resolve, reject){
      try {
        // Prefer mermaid.render if available (Mermaid v10)
        if (mermaid && typeof mermaid.render === 'function') {
          var res1 = mermaid.render(id, code);
          if (res1 && typeof res1.then === 'function') {
            res1.then(function(out){
              try {
                if (out && out.svg) {
                  node.innerHTML = out.svg;
                  if (typeof out.bindFunctions === 'function') out.bindFunctions(node);
                }
                resolve();
              } catch(e1){ reject(e1); }
            }).catch(reject);
            return;
          } else if (res1 && typeof res1.svg === 'string') {
            node.innerHTML = res1.svg;
            if (typeof res1.bindFunctions === 'function') res1.bindFunctions(node);
            resolve();
            return;
          }
        }

        // Fallback to mermaid.mermaidAPI.render (v8/v9)
        if (mermaid && mermaid.mermaidAPI && typeof mermaid.mermaidAPI.render === 'function') {
          var cbHandled = false;
          var res2 = mermaid.mermaidAPI.render(id, code, function(svgCode, bindFns){
            try {
              cbHandled = true;
              if (svgCode) {
                node.innerHTML = svgCode;
                if (typeof bindFns === 'function') bindFns(node);
              }
              resolve();
            } catch(e2){ reject(e2); }
          } /* do not pass container to avoid doc-related issues */);

          if (res2 && typeof res2.then === 'function') {
            res2.then(function(out){
              try {
                if (out && out.svg) {
                  node.innerHTML = out.svg;
                  if (typeof out.bindFunctions === 'function') out.bindFunctions(node);
                }
                resolve();
              } catch(e3){ reject(e3); }
            }).catch(reject);
          } else if (typeof res2 === 'string') {
            node.innerHTML = res2;
            resolve();
          } else if (!cbHandled) {
            // Neither promise nor string nor callback? Treat as error
            reject(new Error('Unexpected return from mermaidAPI.render'));
          }
          return;
        }

        reject(new Error('No supported Mermaid render API'));
      } catch(e){ reject(e); }
    });
  }

  async function renderAll(){
    var theme = computeTheme();
    if (!window.mermaid) return;
    try {
      if (!window.__goGripMermaidInitDone) {
        mermaid.initialize({ startOnLoad: false, theme: theme, securityLevel: 'loose', logLevel: 'error' });
        window.__goGripMermaidInitDone = true;
      } else {
        // Update theme dynamically if needed
        if (mermaid && mermaid.initialize) {
          mermaid.initialize({ startOnLoad: false, theme: theme, securityLevel: 'loose', logLevel: 'error' });
        }
      }
    } catch(e) {
      console.error('Mermaid init error:', e);
    }

    var nodes = Array.prototype.slice.call(document.querySelectorAll('.mermaid'));

    // Sequential rendering to avoid race conditions / global state issues
    for (var i = 0; i < nodes.length; i++) {
      var node = nodes[i];
      try {
        var code = getCode(node).trim();
        setCode(node, code);
        if (mermaid.mermaidAPI && mermaid.mermaidAPI.parse) {
          try {
            mermaid.mermaidAPI.parse(code);
          } catch (parseErr) {
            console.error('Mermaid parse error:', parseErr, { index: i, code: code });
            node.classList.add('mermaid-error');
            node.innerHTML = '<pre class=\"mermaid-error\">Mermaid parse error:\\n' + escapeHTML(parseErr.str || parseErr.message || String(parseErr)) + '</pre>';
            continue;
          }
        }

        var id = 'mermaid-svg-' + i + '-' + Date.now();
        try {
          await renderWithAPI(id, code, node);
          node.classList.remove('mermaid-error');
        } catch(runErr){
          console.error('Mermaid render error:', runErr, { index: i, code: code });
          node.classList.add('mermaid-error');
          node.innerHTML = '<pre class=\"mermaid-error\">Mermaid render error:\\n' + escapeHTML(runErr && runErr.message || String(runErr)) + '</pre>';
        }
      } catch(err){
        console.error('Mermaid error:', err, { index: i });
        node.classList.add('mermaid-error');
        node.innerHTML = '<pre class=\"mermaid-error\">Mermaid error:\\n' + escapeHTML(err.message || String(err)) + '</pre>';
      }
    }
  }

  document.addEventListener('DOMContentLoaded', function(){
    renderAll();

    // Re-render on theme change when in auto mode
    var themeAttr = document.body.getAttribute('data-theme') || 'auto';
    if (themeAttr === 'auto' && window.matchMedia) {
      var mq = window.matchMedia('(prefers-color-scheme: dark)');
      if (mq && mq.addEventListener) {
        mq.addEventListener('change', function(){
          renderAll();
        });
      }
    }
  });
})();