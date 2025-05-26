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

## :question: About

**go-grip** is a lightweight, Go-based tool designed to render Markdown files locally, replicating GitHub's style. It offers features like syntax highlighting, dark mode, and support for mermaid diagrams, providing a seamless and visually consistent way to preview Markdown files in your browser.

This project is a reimplementation of the original Python-based [grip](https://github.com/joeyespo/grip), which uses GitHub's web API for rendering. By eliminating the reliance on external APIs, go-grip delivers similar functionality while being fully self-contained, faster, and more secure - perfect for offline use or privacy-conscious users.

## :zap: Features

- :zap: Written in Go :+1:
- 📄 Render markdown to HTML and view it in your browser
- 📱 Dark and light theme
- 🎨 Syntax highlighting for code
- [x] Todo list like the one on GitHub
- Support for github markdown emojis :+1: :bowtie:
- Support for mermaid diagrams

```mermaid
graph TD;
      A-->B;
      A-->C;
      B-->D;
      C-->D;
```

> [!TIP]
> Support of blockquotes (note, tip, important, warning and caution) [see here](https://github.com/orgs/community/discussions/16925)

> [!IMPORTANT]
>
> test

## :rocket: Getting started

To install go-grip, simply:

```bash
go install github.com/chrishrb/go-grip@latest
```

> [!TIP]
> You can also use nix flakes to install this plugin.
> More useful information [here](https://nixos.wiki/wiki/Flakes).

## :hammer: Usage

By default, go-grip generates static HTML files from your markdown. To render a markdown file:

```bash
go-grip README.md
```

The generated HTML will be saved in your cache directory by default, and the browser will automatically open to view it. You can specify a custom output directory:

```bash
go-grip -o /path/to/output README.md
```

To run as a server instead of generating static files:

```bash
go-grip --server README.md
```

You can also specify a port for the server:

```bash
go-grip --server -p 80 README.md
```

To disable automatic browser opening:

```bash
go-grip -b=false README.md
```

To activate dark mode:

```bash
go-grip --theme dark README.md
```

## :pencil: Examples

<img src="./.github/docs/example-1.png" alt="examples" width="1000"/>

## :bug: Known TODOs / Bugs

- [ ] Tests and refactoring
- [ ] Make it possible to export the generated html

## :pushpin: Similar tools

This tool is a Go-based reimplementation of the original [grip](https://github.com/joeyespo/grip), offering the same functionality without relying on GitHub's web API.
