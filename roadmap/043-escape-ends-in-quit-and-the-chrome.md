# 43. Escape ends in quit, and the chrome answers only the keyboard

**Done 2026-09-26.**

**Escape:** the last rung opens settings instead of arming the quit prompt. Quit stays on `q`, which is what the footer always said.

**Mouse:** the inventory came back much shorter than this entry assumed. Six surfaces already answered a click — the view key, timeline, breakdown, hover card, sidebar and strip. Only the footer, settings, search and help did not. All four now do.

**How a click runs the same code path rather than a second one: it presses the key.** `Game.just` reads a `clicked` key alongside the real keyboard, so a click on a verb sets it and the keyboard's own handler runs in the same frame. Settings does it twice over — the click puts the cursor on the row under the pointer, then presses Enter.

That keeps Aria's reason intact rather than merely satisfying it: a verb goes on showing its key *because the key is what the click presses*, so there is no second implementation to drift out of step with the label.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
