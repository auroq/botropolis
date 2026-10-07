# quillmark

A very small static-site generator.
Markdown goes in, HTML comes out, and nothing outside the Python standard library is involved.

## Usage

```sh
python -m quillmark new "Fog signals"   # writes content/fog-signals.md
python -m quillmark build               # renders content/ into public/
```

`--root DIR` points either command at a site other than the current directory.
`build --out DIR` writes somewhere other than `public/`.

## Layout

| path | holds |
| --- | --- |
| `site.toml` | site title, base URL and default author |
| `content/*.md` | one post per file, each with a front-matter block |
| `templates/post.html` | the page for one post |
| `templates/index.html` | the list of posts, newest first |
| `public/` | the output, safe to delete |

## Front matter

Each post starts with a block of `key: value` lines between `---` fences.

```markdown
---
title: Reading tide tables by hand
date: 2026-09-02
author: M. Trevose
---
```

`title` falls back to the first `# heading`, and then to the file name.
`slug` overrides the output file name.
Posts are sorted by `date` as written, so use ISO dates.

## Markdown

quillmark understands a deliberate subset: headings, paragraphs, strong and emphasis, code spans, links, fenced code blocks, and flat lists.
Anything it does not recognise is kept as paragraph text.

## Development

```sh
python -m venv .venv && .venv/bin/pip install -e '.[test]'
.venv/bin/python -m pytest -q
```
