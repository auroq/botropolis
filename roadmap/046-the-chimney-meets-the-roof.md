# 46. The chimney meets the roof on a straight line, and a cylinder cannot

**Done 2026-09-26 (r226).**

**The geometry is the whole bug:** a cylinder cut by a plane meets it in an ellipse, never a straight line.

The stack is sunk so the roof plane cuts it, and a flashing ellipse drawn over the join becomes the round bottom line, with a collar band above it and braces below.

**Only the near half of each ring is drawn.** The first cut filled them as whole ellipses, which painted discs over the chimney and left a sliver of it showing between two grey pancakes.

**Two braces, not three.** A stack carries three at 120°, but at 270° in plan the third is directly behind the pipe and entirely hidden, so drawing it would paint nothing. `MinStackBracePx = 12` on the stack's drawn width.

Frame `docs/screenshots/r226-chimney-join.png`, the same view before and after.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
