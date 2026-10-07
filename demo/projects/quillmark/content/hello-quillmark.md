---
title: Hello, quillmark
date: 2026-08-14
---

This site is built by **quillmark**, a generator small enough to read in one sitting.
It has no dependencies beyond the standard library.

## Why another one

Most generators do a great deal.
This one does *very little*, on purpose:

- reads `content/*.md`
- renders `templates/` with `string.Template`
- writes `public/`

The source lives on [the project page](https://notes.example.org/quillmark).
