---
name: changelog
description: Update CHANGELOG.md with what changed since the last entry. Use after finishing a change, or when asked to write or update the changelog or release notes.
---

# Changelog

Keep `CHANGELOG.md` in the Keep a Changelog format.

1. Read `CHANGELOG.md`; if there is none, create it with a `# Changelog` heading and an `## [Unreleased]` section.
2. Run `git log --oneline` since the newest commit the changelog already describes, and `git diff --stat` for anything uncommitted.
3. Sort each change into `### Added`, `### Changed`, `### Fixed` or `### Removed` under `## [Unreleased]`.
4. Write one line per change, in the past tense, saying what a user of the project would notice rather than which file moved.
5. Mention an issue id such as `TIDE-12` when the change closes one.

Do not invent changes that are not in the history or the diff.
