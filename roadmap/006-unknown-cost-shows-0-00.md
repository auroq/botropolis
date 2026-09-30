# 6. Unknown cost shows `~$0.00`

**Fixed 2026-09-18 (`1446c8b`, `049cc5a`): a known flag rides with the cost; unknown is a dash everywhere.**

The breakdown over the sample fixture shows 398.9M tokens at ~$0.00; the rule that applies to the context window applies here:
no cost-state means `—`, never a number.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
