# MathJax mhchem font extension (CHTML)

Vendored so that `\ce{...}` and `\pu{...}` render offline. The TeX `mhchem`
extension (`defaults/static/js/input/tex/extensions/mhchem.js`) loads
`[fonts]/mathjax-mhchem-font-extension/chtml.js`, which then loads its woff2
font from `chtml/woff2/`. Without these files MathJax fetches them from
cdn.jsdelivr.net.

- Package: `@mathjax/mathjax-mhchem-font-extension` 4.1.0, matching
  `mathjax@4.1.0` (`defaults/static/js/tex-mml-chtml.js`)
- Source: https://registry.npmjs.org/@mathjax/mathjax-mhchem-font-extension/-/mathjax-mhchem-font-extension-4.1.0.tgz
  (sha1 `f4b64fbca4a48e336e394d4588a4d2a17f9a4ae3`)
- Copied unmodified: `package/chtml.js` and `package/chtml/woff2/`
- License: Apache-2.0 (the package's `license` field), see `LICENSE`

`defaults/static/js/mathjax-options.js` points MathJax's `[fonts]` loader path
at `/static/fonts`, so `[fonts]/mathjax-mhchem-font-extension/...` resolves
here. When upgrading MathJax, replace these files with the same version.
