"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

const script = fs.readFileSync(path.join(__dirname, "../static/js/theme-switch.js"), "utf8");
const key = "go-grip-theme";
const nonmatching = "(prefers-color-scheme:" + "/*x*/".repeat(30) + "unknown)";
const queries = [
  ["(prefers-color-scheme: dark)", "(prefers-color-scheme: light)"],
  ["(prefers-color-scheme: light)", "(prefers-color-scheme: dark)"],
  ["(min-width: 600px) and (prefers-color-scheme: dark)", "(min-width: 600px) and (prefers-color-scheme: light)"],
  ["(min-width: 600px)", "(min-width: 600px)"],
  [
    "(prefers-color-scheme: dark), (/*a*/prefers-color-scheme/*b*/:/*c*/light/*d*/)",
    "(prefers-color-scheme: light), (/*a*/prefers-color-scheme/*b*/:/*c*/dark/*d*/)",
  ],
  [nonmatching, nonmatching],
];
const other = (theme) => theme === "dark" ? "light" : "dark";

function load(browser, storage) {
  const sources = queries.map(([media]) => ({ media }));
  const events = {};
  let theme;
  const on = (name, fn) => { events[name] = fn; };
  const button = { querySelector: () => ({}), addEventListener: on };
  const preference = { get matches() { return browser === "dark"; }, addEventListener: on };
  vm.runInNewContext(script, {
    window: { matchMedia: () => preference },
    localStorage: {
      getItem: (name) => storage[name],
      setItem: (name, value) => { storage[name] = value; },
    },
    CustomEvent: function () {},
    document: {
      body: {
        setAttribute: (_, value) => { theme = value; },
        getAttribute: () => theme,
        dispatchEvent: () => {},
      },
      getElementById: (id) => id === "theme-toggle" ? button : null,
      querySelectorAll: (selector) => {
        assert.equal(selector, "picture > source[media]");
        return sources;
      },
      addEventListener: on,
    },
  });
  events.DOMContentLoaded();
  return {
    sources,
    toggle: () => events.click(),
    setBrowser: (value) => { browser = value; events.change(); },
    check: (expected) => {
      assert.equal(theme, expected);
      assert.deepEqual(sources.map((source) => source.media), queries.map((pair) => pair[Number(browser !== expected)]));
    },
  };
}

for (const browser of ["light", "dark"]) {
  const storage = {};
  load(browser, storage).check(browser);
  for (const preview of ["light", "dark"]) {
    storage[key] = preview;
    const page = load(browser, storage);
    page.check(preview);
    for (const expected of [other(preview), preview, other(preview), preview]) {
      page.toggle();
      page.check(expected);
    }
    page.setBrowser(other(browser));
    page.check(preview);
    load(other(browser), storage).check(preview);
    page.setBrowser(browser);
    page.check(preview);
    page.sources[3].media = "(min-width: 900px)";
    page.toggle();
    assert.equal(page.sources[3].media, "(min-width: 900px)");
    page.setBrowser(other(browser));
    assert.equal(page.sources[3].media, "(min-width: 900px)");
  }
}
