# 6. Unknown cost shows `~$0.00`

**Fixed 2026-09-18 (`73e13ea`, `1d6a787`): a known flag rides with the cost; unknown is a dash everywhere.**

The breakdown over the sample fixture shows 398.9M tokens at ~$0.00; the rule that applies to the context window applies here:
no cost-state means `—`, never a number.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
