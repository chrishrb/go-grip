# MathJax TeX extensions

Vendored so that math renders offline. The combined bundle
`defaults/static/js/tex-mml-chtml.js` (MathJax 4.1.0) loads TeX extensions
that it does not contain from `[tex]` = `[mathjax]/input/tex/extensions`,
which resolves to `/static/js/input/tex/extensions/`. It does so when a macro
from its autoload list is first used (`\boldsymbol`, `\cancel`, `\ce`,
`\braket`, `\color`, ...), for `\require{...}`, and for the packages that
`defaults/static/js/mathjax-options.js` preloads to match github.com
(`mathtools`, `upgreek`, `gensymb`, ...).

- Package: `mathjax` 4.1.0, the same version as `tex-mml-chtml.js`
- Source: https://registry.npmjs.org/mathjax/-/mathjax-4.1.0.tgz
  (sha1 `71036219600120b24409faf2514edd3d73ff4ddc`)
- Copied unmodified: every file in `package/input/tex/extensions/` except
  `bbm.js`, `bboldx.js` and `dsfont.js`. Those need the
  `@mathjax/mathjax-{bbm,bboldx,dsfont}-font-extension` packages (about 1.7 MB
  together), which are not vendored, so `mathjax-options.js` disallows them
  in `\require`. GitHub does not support them either.
- License: Apache-2.0 (the package's `license` field), see `LICENSE`

`mhchem.js` (`\ce`, `\pu`) also loads its font extension, vendored under
`defaults/static/fonts/mathjax-mhchem-font-extension/`.

`defaults/embed_test.go` checks that every extension the bundle can load is
vendored here or disallowed. When upgrading MathJax, replace these files with
the same version's `input/tex/extensions/`.
