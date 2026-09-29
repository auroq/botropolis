# 19. Rovers only bob

**Fixed 2026-09-21 (r140). A worker that stands still does not read as working, and motion should mean something: the rover drives from the door to the block's avenue edge and back once per tool call (`PreToolUse` out, `PostToolUse` back, along the district's own cells), so every trip is a tool call you can count; idle between calls it waits at the door; under `reduced_motion` it stays at the door.**

`city.Trip` is the leg, `District.Route` the way out — the door, the service lane beside the building, the kerb — and a call that ends early turns the rover round where it stands.
Reduced motion got its own switch on the scene on the way: a still frame sets `SetInstant` so the camera does not have to ease into place, which is not the same as asking everything to stand still,
and conflating them meant a screenshot could never show a worker anywhere but its door.
Frame `docs/screenshots/r140-rover-trip.png`.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
