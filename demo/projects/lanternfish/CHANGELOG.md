# Changelog

## [Unreleased]

## [0.2.0] - 2026-09-24

### Added

- `GET /lanterns/{id}` returns a single lantern.
- `-seed` loads a starting inventory when there is no snapshot.

### Fixed

- Adding a lantern that claims to be checked out stores it as available.

## [0.1.0] - 2026-09-05

### Added

- List, add, check out and return lanterns over HTTP.
- The inventory is saved to a JSON snapshot after every change.
