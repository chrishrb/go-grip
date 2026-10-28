# MathJax New Computer Modern font (CHTML)

Vendored so that math renders offline. Without these files MathJax fetches
its fonts and dynamic font data from cdn.jsdelivr.net.

- Package: `@mathjax/mathjax-newcm-font` 4.1.0, the version that
  `mathjax@4.1.0` (`defaults/static/js/tex-mml-chtml.js`) depends on
- Source: https://registry.npmjs.org/@mathjax/mathjax-newcm-font/-/mathjax-newcm-font-4.1.0.tgz
  (sha1 `746a03368eb5fa611b2aac46de26277bdedada12`)
- Copied unmodified: `package/chtml/woff2/` and `package/chtml/dynamic/`
- License: Apache-2.0 (the package's `license` field), see `LICENSE`

`defaults/static/js/mathjax-options.js` points MathJax's `[fonts]` loader path
at `/static/fonts`, so `[fonts]/mathjax-newcm-font/chtml/...` resolves here.
When upgrading MathJax, replace these files with the font version its
`package.json` depends on.
