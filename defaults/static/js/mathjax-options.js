MathJax = {
  // MathJax resolves its font files (woff2) and the dynamically loaded font
  // data for less common characters under the [fonts] path, which defaults to
  // cdn.jsdelivr.net. Point it at the vendored copy so math renders offline.
  loader: {
    paths: {
      fonts: '/static/fonts',
    },
    // github.com renders math with these TeX packages enabled up front. The
    // bundle autoloads most of them on first use from the vendored
    // static/js/input/tex/extensions/, but these define macros it cannot
    // autoload (\coloneqq, \upalpha, \degree, \centernot, ...), so load
    // them with the page, as GitHub does.
    load: [
      '[tex]/cases',
      '[tex]/centernot',
      '[tex]/empheq',
      '[tex]/gensymb',
      '[tex]/mathtools',
      '[tex]/textcomp',
      '[tex]/upgreek',
    ],
  },
  tex: {
    packages: {
      '[+]': ['cases', 'centernot', 'empheq', 'gensymb', 'mathtools', 'textcomp', 'upgreek'],
    },
    require: {
      // These extensions need font extensions that are not vendored, so
      // \require{...} would fail to load their fonts. GitHub does not
      // support them either.
      allow: {
        bbm: false,
        bboldx: false,
        dsfont: false,
      },
    },
  },
  options: {
    a11y: {
      backgroundOpacity: 0,
    },
    enableMenu: false,
    // MathJax 4 generates speech and Braille in a web worker that loads
    // static/js/sre/speech-worker.js plus its rule maps. Those are not
    // vendored, so the worker only ever failed with a console error. Turn
    // speech and Braille off so it is never requested. The menu settings win
    // over the document options, so both are set.
    enableSpeech: false,
    enableBraille: false,
    menuOptions: {
      settings: {
        speech: false,
        braille: false,
      },
    },
  }
}
