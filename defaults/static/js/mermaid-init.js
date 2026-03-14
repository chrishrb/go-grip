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
