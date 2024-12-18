<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="#">
    <img src=".github/docs/logo-1.png" alt="Logo" height="120">
  </a>

  <h3 align="center">go-grip</h3>

  <p align="center">
    Render your markdown files local<br>- with the look of GitHub
  </p>
</div>

## ⚡️Features

* ⚡️Written in Go :+1:
* 📄 Render markdown to HTML and view it in your browser
* 📱 Dark and white mode
* 🎨 Syntax highlighting for code 
* [x] Todo list like the one on GitHub
* Support for github markdown emojis :+1: :bowtie:
* Support for mermaid diagrams

```mermaid
graph TD;
      A-->B;
      A-->C;
      B-->D;
      C-->D;
```

> [!TIP]
> Support of blockquotes (note, tip, important, warning and caution) [see here](https://github.com/orgs/community/discussions/16925)


## 🚀 Getting started

To install go-grip, simply:

```bash
go install github.com/chrishrb/go-grip@latest
```

> [!TIP]
> You can also use nix flakes to install this plugin.
> More useful information [here](https://nixos.wiki/wiki/Flakes).

## 🔨 Usage

To render the `README.md` file simply execute:

```bash
go-grip README.md
# or
go-grip
```

The browser will automatically open on http://localhost:6419. You can disable this behaviour with the `-b=false` option.

You can also specify a port:

```bash
go-grip -p 80 README.md
```

or just open a file-tree with all available files in the current directory:

```bash
go-grip -r=false
```

It's also possible to activate the darkmode:

```bash
go-grip -d .
```

To terminate the current server simply press `CTRL-C`.

## 📝 Examples

<img src="./.github/docs/example-1.png" alt="examples" width="1000"/>

## 🐛 Known TODOs / Bugs

* [ ] Tests and refactoring
* [ ] Make it possible to export the generated html

## 📌 Similar tools

This tool is a Go-based reimplementation of the original [grip](https://github.com/joeyespo/grip), offering the same functionality without relying on GitHub's web API.
