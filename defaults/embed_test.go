package defaults

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

const mathJaxFontDir = "static/fonts/mathjax-newcm-font/chtml"

// TestMathJaxFontsVendored guards against MathJax falling back to its CDN for
// fonts: every woff2 file the bundled MathJax names must be embedded, and the
// options must point MathJax's [fonts] path at the embedded copy.
func TestMathJaxFontsVendored(t *testing.T) {
	t.Parallel()

	bundle, err := fs.ReadFile(StaticFiles, "static/js/tex-mml-chtml.js")
	if err != nil {
		t.Fatalf("read MathJax bundle: %v", err)
	}
	refs := regexp.MustCompile(`mjx-ncm-[a-z0-9-]+\.woff2`).FindAllString(string(bundle), -1)
	if len(refs) == 0 {
		t.Fatal("expected the MathJax bundle to reference mjx-ncm-*.woff2 fonts")
	}
	for _, name := range refs {
		if _, err := fs.Stat(StaticFiles, mathJaxFontDir+"/woff2/"+name); err != nil {
			t.Errorf("font %s referenced by the MathJax bundle is not embedded: %v", name, err)
		}
	}

	dynamic, err := fs.ReadDir(StaticFiles, mathJaxFontDir+"/dynamic")
	if err != nil {
		t.Fatalf("read dynamic font data dir: %v", err)
	}
	if len(dynamic) == 0 {
		t.Fatal("expected dynamic font data files to be embedded")
	}

	options, err := fs.ReadFile(StaticFiles, "static/js/mathjax-options.js")
	if err != nil {
		t.Fatalf("read MathJax options: %v", err)
	}
	if !strings.Contains(string(options), "fonts: '/static/fonts'") {
		t.Fatalf("expected mathjax-options.js to set loader.paths.fonts to /static/fonts, got:\n%s", options)
	}
}

const mathJaxTexExtDir = "static/js/input/tex/extensions"

// TestMathJaxTexExtensionsVendored guards against 404s for TeX extensions the
// bundled MathJax loads on demand (autoload, \require, loader.load) from
// [tex] = /static/js/input/tex/extensions. Every extension the bundle knows
// must be vendored there or disallowed for \require in mathjax-options.js,
// and every vendored extension's font extension must be vendored too.
func TestMathJaxTexExtensionsVendored(t *testing.T) {
	t.Parallel()

	bundle, err := fs.ReadFile(StaticFiles, "static/js/tex-mml-chtml.js")
	if err != nil {
		t.Fatalf("read MathJax bundle: %v", err)
	}
	options, err := fs.ReadFile(StaticFiles, "static/js/mathjax-options.js")
	if err != nil {
		t.Fatalf("read MathJax options: %v", err)
	}

	// The bundle's loader dependency map names every [tex] component it can
	// load: "[tex]/boldsymbol":["input/tex-base"], ...
	known := map[string]bool{}
	for _, m := range regexp.MustCompile(`"\[tex\]/([a-z0-9]+)":\[`).FindAllStringSubmatch(string(bundle), -1) {
		known[m[1]] = true
	}
	if len(known) < 30 {
		t.Fatalf("expected the MathJax bundle to name the [tex] extensions it can load, found %d", len(known))
	}

	for name := range known {
		_, err := fs.Stat(StaticFiles, mathJaxTexExtDir+"/"+name+".js")
		disallowed := regexp.MustCompile(`\b` + name + `: false\b`).Match(options)
		switch {
		case err == nil && disallowed:
			t.Errorf("TeX extension %s is vendored but disallowed in mathjax-options.js", name)
		case err != nil && !disallowed:
			t.Errorf("TeX extension %s is loadable by the MathJax bundle but neither vendored in %s nor disallowed in mathjax-options.js", name, mathJaxTexExtDir)
		}
	}

	entries, err := fs.ReadDir(StaticFiles, mathJaxTexExtDir)
	if err != nil {
		t.Fatalf("read TeX extension dir: %v", err)
	}
	fontExt := regexp.MustCompile(`"\[tex\]/[a-z0-9]+","(mathjax-[a-z0-9]+-font-extension)"`)
	fontExts := 0
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".js")
		if !ok {
			continue
		}
		if !known[name] {
			t.Errorf("vendored TeX extension %s is not known to the MathJax bundle; stale file?", e.Name())
		}
		src, err := fs.ReadFile(StaticFiles, mathJaxTexExtDir+"/"+e.Name())
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		for _, m := range fontExt.FindAllSubmatch(src, -1) {
			fontExts++
			dir := "static/fonts/" + string(m[1])
			if _, err := fs.Stat(StaticFiles, dir+"/chtml.js"); err != nil {
				t.Errorf("TeX extension %s needs font extension %s, which is not embedded: %v", name, m[1], err)
			}
			woff2, err := fs.ReadDir(StaticFiles, dir+"/chtml/woff2")
			if err != nil || len(woff2) == 0 {
				t.Errorf("TeX extension %s needs the woff2 fonts of %s, which are not embedded: %v", name, m[1], err)
			}
		}
	}

	if fontExts == 0 {
		t.Error("expected a vendored TeX extension (mhchem) to load a font extension; has the bundle format changed?")
	}

	preloads := regexp.MustCompile(`'\[tex\]/([a-z0-9]+)'`).FindAllStringSubmatch(string(options), -1)
	if len(preloads) == 0 {
		t.Fatal("expected mathjax-options.js to preload the TeX packages github.com enables")
	}
	for _, m := range preloads {
		if _, err := fs.Stat(StaticFiles, mathJaxTexExtDir+"/"+m[1]+".js"); err != nil {
			t.Errorf("mathjax-options.js loads [tex]/%s, which is not vendored: %v", m[1], err)
		}
	}
}
