# Botropolis — roadmap

The plan and its current state.
Every numbered item has its own file in [`roadmap/`](roadmap/); this file is the index and nothing else.

Detail belongs in `roadmap/`, rules and what-exists belong in [DESIGN.md](DESIGN.md), and neither belongs here.
This file grew to 2,689 lines by holding both, which is what it is being kept short to avoid.

## Phases

| phase | what | state |
| --- | --- | --- |
| 7 | UI foundation | shipped 2026-09-18 |
| 8 | City plan | shipped 2026-09-18 |
| 9 | Art pipeline and art pass | shipped 2026-09-18 |
| 10 | Parity with bot-crossing | shipped 2026-09-18 |
| 11 | Info and control | shipped 2026-09-18 |
| 12 | Packaging and polish | shipped 2026-09-18 |
| 13–18 | audit follow-up | done, [`audit-r96.md`](roadmap/audit-r96.md) |
| 19 | Info views | done, [`phase-19-info-views.md`](roadmap/phase-19-info-views.md) |
| 20 | The map has to explain itself | one item open, [`phase-20-explain-itself.md`](roadmap/phase-20-explain-itself.md) |
| 21 | The art batch | done, items 35–68 |

Phases 7–12 were specified in [`phases-7-12.md`](roadmap/phases-7-12.md) and audited at r96.

## Open

| item | what it wants | whose call |
| --- | --- | --- |
| [58](roadmap/058-repo-pack-size-and-growth.md) | the repo pack, priced three ways with **C — leave it** recommended | Aria |
| [66](roadmap/066-115-mb-private-dirty-curiosity.md) | ~115 MB of private dirty unaccounted for, filed as a curiosity | nobody, until it costs something |
| [phase 20 item 3](roadmap/phase-20-explain-itself.md) | the plant's *reading* — the geometry was fixed at r186 and has not been looked at since | Aria, from a frame |

## Waiting on Aria

- **The plant's reading.** Wants a frame, not a measurement.
- **Item 58.** The recommendation is to leave the pack alone; the ruling is hers.
- **The validation checklist.** Ten steps, in [`inventory-2026-09-29.md`](roadmap/inventory-2026-09-29.md), never run end to end.
- **The nine unadjudicated entries** — items 22, 23, 24, 26, 27, 28, 29, 31 and 33.
  Each was filed as a measurement or a finding rather than a bug, or was superseded by a later item, and none was ever formally closed.
  Several are visibly superseded downstream — the CPU bar by item 60, the undrawn atlas by item 24's option D and item 34, the memory question by item 61 — but saying which are closed is an afternoon of careful reading and has not been done.
  Until it is, this list is what is *tracked* open, not a proof that nothing else is.

## Not doing

Planets, orbit mode, a 3D renderer, network serving, quality presets beyond render scale and reduced motion, animated faces.
Each is either bot-crossing's setting rather than a feature, or a cost the footprint principle rules out.

## Where things are

| | |
| --- | --- |
| `roadmap/NNN-*.md` | one file per numbered item, 1–68 |
| `roadmap/phase-*.md` | the phases that carry their own sub-items |
| `roadmap/assets/` | references and crops belonging to an item |
| `docs/screenshots/` | frames, cited by revision |
| `docs/references/` | annotated crops from investigations |
| [DESIGN.md](DESIGN.md) | what exists, and the rules the work is held to |

Items are numbered in one series, 1–68.
Section 8's bugs 1–34 and the art batch's 35–68 continue the same numbering; phases 19 and 20 have their own small series, cited as "phase 20 item 3".
There are two bug 18s — road tiles, and rovers parking — both fixed, both left renumbered as written, because a dated record that gets tidied stops being a record.
Nothing cites either by number.

## Status, 2026-09-29

68 items, 57 closed, 2 open, 9 unadjudicated — counted by number.
By *file* it is 58 closed of 69, because the two bug 18s are one number and two entries.
Both figures are the same fact; this line quotes the first.
Gate green: `go build`, `go vet`, 22 test packages, `make lint` 0 issues.
The inventory is [`inventory-2026-09-29.md`](roadmap/inventory-2026-09-29.md).
