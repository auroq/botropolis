# 58. The repo pack is 157 MB, grows by about 27 MB per atlas re-cut, and the obvious remedy is closed off by `pkgver()`

**Ruled 2026-09-29 by Aria: leave it.** The pack stays as it is. The trigger to revisit is the repo going public, not the pack passing a size — and that is also the last cheap moment for a history rewrite.

Measured 2026-09-26 at r268: `.git` is **157 MB**, and 154 blob versions of `pkg/assets/kits` account for **248 MiB raw** across 34 commits that touched the atlas.
PNGs are already compressed, so they do not delta against each other — every re-cut adds the whole atlas again, and the current atlas is 27 MB.
The roadmap's own note says "the repo pack is 30 MB, almost all atlases", which was true on 2026-09-21 and is now off by a factor of five.

The trap is in the shape of the cheap fix.
`PKGBUILD` clones the full repo and `pkgver()` runs `git rev-list --count HEAD` to produce the `r268` in every package name Aria installs, so **a shallow clone cannot be used** — it would take the revision number with it.
`makepkg` therefore transfers 155 MB to build a 55 MB binary, and that grows monotonically with every art change.

Three real options, and this is Aria's call because each trades something different:
git-lfs for `pkg/assets/kits` (was ruled out on 2026-09-21 on the grounds that the atlas does not churn — the ground has moved, it has been re-cut ~~34~~ 18 times — see the correction below);
build the atlases in the PKGBUILD (needs Blender as a makedepend, which is a heavy dependency for a package that exists to be installed);
or leave it, on the grounds that one machine's clone is one machine's disk.
Note that whatever is chosen, only history rewriting recovers the 157 MB already spent — the decision is about the next ~~34~~ 18 re-cuts.

**Priced 2026-09-26 at r272. Recommendation: leave it, and spend the free hygiene first. The decision is Aria's; this is the case, not the ruling.**

Three of the numbers this was filed on need correcting before the options are worth comparing.

**`.git` is 146 MB, not 157 MB, and 21 MB of the difference is slack that `git gc` reclaims for nothing.**
The 157 MB reading was taken with 930 loose objects outstanding — 85 MiB of them — which is what an un-gc'd repo looks like mid-session, not its steady state.
After `git gc`: 146 MB on disk, a 144.7 MiB pack, no loose objects.

**18 commits have touched `pkg/assets/kits`, not 34.**
26 have touched `pkg/assets` if the Kenney tilemaps are counted with the atlas.
Neither is 34, and I cannot reconstruct where 34 came from, so the per-re-cut arithmetic downstream of it should be redone rather than trusted.

**Where 34 came from, since it is reconstructible and the cause has now cost two bugs in one session.**
This repo sets `log.showSignature=true`, so every commit `git log --oneline` prints carries a second line — `Good "git" signature for dev@ariavesta.com …`.
`git log --oneline -- pkg/assets/kits | wc -l` therefore counts two lines per commit: 36 today for the 18 commits, and 34 when I ran it at r268 for 17.
So the figure was not a wrong count of the right thing, it was **a line count read as a commit count** — the same shape as the atlas cell read as a city tile, and it is the third time this project has been bitten by a number in the wrong units.
The same setting is why the AUR `Makefile`'s `commit` target wrote a mangled subject: `$(git log -1 --format=%s)` returned the verification line ahead of the message.
**In this repo any script reading git log must pass `--no-show-signature`, and `git rev-list --count` is safe because it prints no signatures** — which is why `pkgver()` was never affected.
The blob count is right: 160 versions at r272, which is the filed 154 plus the six files item 59's re-cut rewrote.

**PNGs do not delta, but they do deflate: 256 MiB raw becomes 91.6 MiB in the pack, a factor of 2.8.**
That is worth knowing in itself — it says the shipped PNGs are not compressed as hard as they could be, since a general-purpose deflate still finds two thirds of them redundant, and `tools/shrink-pngs` exists and is not in the pipeline.
Of the 144.6 MiB pack, the atlas is **91.6 MiB (63%)** and everything else is 53.0 MiB.

**Growth is 8–27 MB per re-cut, not a flat 27.**
Item 59's re-cut rewrote 6 of 11 files for 8.2 MiB raw, because the shelf packer is deterministic and only repacks from the first changed piece onward — `kits-z1-0` and `kits-z2-0`–`3` came out byte-identical.
A change to an early piece rewrites all nine pages and does cost the full 27 MB.

**And `makepkg` transfers 155 MB once per machine, not once per build.**
It keeps a bare cache clone beside the PKGBUILD and fetches into it; the working copy in `src/` is a local checkout.
So the recurring network cost of a build is the fetch delta — 8 MB for item 59 — and the recurring *disk* cost is two 155 MB copies that never shrink.

#### The options, priced

**A — git-lfs for `pkg/assets/kits`.** Works, and cheaper to adopt than the 2026-09-21 ruling assumed, but it buys the wrong thing.
It does not shrink the pack by one byte; it stops the pack growing, and the 91.6 MiB already spent stays spent unless history is rewritten.
`git-lfs` is not installed on this machine and becomes a hard makedepend, because `makepkg` checks out a working copy and the smudge filter is what turns pointers back into PNGs.
The failure mode if it is missing is better than feared: `check()` runs `go test ./cmd/... ./pkg/...`, so a checkout of pointer files fails the atlas tests at build time rather than shipping a city with no sprites.
A contributor without `git-lfs` gets 130-byte text files where the PNGs should be and a build that fails in `pkg/assets`.
The real cost is the quota: GitHub gives 1 GB of LFS storage and 1 GB/month of bandwidth free, every version of every atlas counts against storage, and at 8–27 MB a re-cut the free tier is roughly 35 re-cuts away.
That trades a disk problem nobody has for a metered bill, on one machine's private repo.

**B — build the atlases in the PKGBUILD. Disqualified, and not on cost.**
`pkg/assets/kits.go` embeds them with `//go:embed kits/*.png`, and a Go embed pattern that matches no files is a compile error, verified: `pattern kits/*.png: no matching files found`.
So taking the atlas out of git does not make the repo lighter to build — it makes the repo impossible to build, for everyone, until Blender has run.
`go test`, `go build` and every editor's language server stop working on a fresh clone.
The cost is real too: `blender` is 384.75 MiB installed behind a 418-package dependency tree, and the re-cut is about 100 s per zoom against a package build that takes 22 s today.
But the dependency is the second reason to refuse this, not the first.

**C — leave it. Recommended.**
The growth is 8–27 MB per re-cut on one machine's disk, and the repo is not close to being the biggest thing in this workflow.

#### The number that actually wants spending

`~/workspaces/aur/botropolis-git` is **2.9 GB**, and **134 accumulated `.pkg.tar.zst` files are 2.68 GB of it** — every package built since r21, none ever removed.
That is eighteen times the entire repository, sitting next to it, reclaimed by the `clean` target item 57 already wrote.
Optimising a 146 MB repo while 2.68 GB of stale build output sits in the adjacent directory is fixing the smallest number in view, which is the shape item 36 and item 52 both had: a measurement that is correct and points at the wrong thing.

So: `git gc` (−21 MB, done), `make clean` in the AUR repo (−2.68 GB, offered and not taken unilaterally), and leave the pack alone.

**Aria ran `make clean` herself on 2026-09-29 at 10:09.**
The directory is 202 MB where it was 2.9 GB, and what is left is the r293 package pair and the bare cache clone.
The recommendation on the repo itself — option C, leave it — is unchanged and still hers to rule on.

**The trigger to revisit is the repo going public**, not the pack passing a size.
While it is one person's clone, 146 MB is a rounding error against the 2.9 GB beside it.
The moment anyone else clones it, the cost becomes theirs and recurring, and that is also the last moment a history rewrite is cheap — before there are clones to invalidate.
LFS with `git lfs migrate import` done once at that point is the right shape; LFS bolted on now is the same work minus the saving.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
