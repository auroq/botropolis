# 67. The usage probe's transcripts are never pruned, and lingering sessions are what this project is for

**Done 2026-09-26 (`2708cc4`, r290).**

Measured 2026-09-26 at r286: `~/.claude/projects/-home-avesta--local-state-botropolis-usage-probe/` holds **7 transcripts**, one per refresh since 18:42, about 4 KB each.
Nothing deletes them. `RefreshUtilization` creates the directory and runs `claude -p "/usage"` in it; the fence (item 54) is the whole of the design, and the fence is about where they land rather than whether they stay.

**The fence works and that part is verified**: every transcript containing `"/usage"` on this machine is inside that folder, and there are none anywhere else, so the city is clean and the 22 strays Aria had me delete have not come back.
The build session's standing list still names those 22 as outstanding; they were deleted, and this is what replaced them.

Why it is worth a line at all, given it is 28 KB: **bot-crossing existed because sessions lingered, and this is botropolis leaving litter of exactly that kind in exactly that directory.**
One transcript per keypress, kept forever, in the folder the tool tells the loader to ignore — ignored is not the same as tidy, and the next person to run `ls ~/.claude/projects` will find a folder that only grows.

The fix is small and the choice is about what "enough" means: delete the probe's transcript after reading the figures, since the file has served its purpose the moment `~/.claude.json` is re-read; or keep the newest and drop the rest, if one is worth having to debug a failed refresh.
I would keep one. A probe that leaves no trace at all is a probe you cannot ask why it failed.

**Fixed, keeping one, for the reason given.** `pruneProbeTranscripts` runs at the end of `RefreshUtilization`, after the probe rather than before, so the transcript that survives is the one the current run just wrote — which is the one worth having when the figures come back wrong.

Nine by the time I got to it, not seven; two more arrived while the entry was being written, which is the growth rate the entry describes making its own case.

Three decisions in it worth stating, because each is a way this could have been quietly wrong:

- **Newest by modification time, not by name.** Claude Code names transcripts with a UUID, which sorts arbitrarily. Sorting by name would look ordered, run without error, and keep whichever run happened to sort last — a wrong answer with no symptom, which is this project's most expensive shape.
- **Only `*.jsonl` is touched.** This deletes from the user's `~/.claude`, so it removes what the probe made and nothing else. A `notes.txt` in that folder survives, and there is a test that says so.
- **Best-effort, and after the result is decided.** A probe that fetched the figures has done its job whether or not the tidying worked, so a failed prune does not turn a successful refresh into a failed one.

A missing folder is not an error either — the probe may never have run — and `keep` is a parameter rather than a constant inside the loop, so the test can ask for the boundary cases directly instead of asserting `probeKeep`'s value twice.

**Four tests**: several runs leaving exactly the newest; a single run left alone; a folder that never existed; and a non-transcript file surviving.

### Worth knowing, not bugs

- ~~**A third of the atlas is never drawn.**~~ **Answered and closed by bug 24; Aria chose option D on 2026-09-25 (r188).**
  Eight commercial pieces went into `fillClasses`, ten orphans left `PIECES`, seven were kept with a stated use.
  The standing question this bullet recorded — reserve or oversight — has an answer per piece and does not need Aria again.
  What is true today is item 59: 8 of 79 undrawn, of which six are that reserve and two are new orphans created since.
- The client binary is **47.8 MB** and the daemon **13.8 MB** (measured 2026-09-26 after item 63; the atlas is 20.7 MB of the client, and the daemon links none of the UI).
  It was 54.9 MB at r268: item 59 took 0.5 MB and item 63 took 6.6 MB, both of them atlas rather than code.
  The 20 MB bar is the daemon's and it holds; the client has never had one.
- The repo pack was 30 MB on 2026-09-21 and is **157 MB** at r268 — see item 58, which is this bullet's concern arriving by the route it did not expect.
  If `make sprites` churns, that is git-lfs or build-time atlases in the PKGBUILD.
  Measured 2026-09-21 with the new `make sprites-check` (phase 15): it does not churn.
  ~~A no-op render reproduces every atlas bit for bit — same manifests, same IDAT bytes, same decoded pixels on all nine pages.~~
  **Overstated; see item 65.** `atlas-diff.py` compares decoded pixels with a tolerance and only the manifests byte for byte, and it says in its own docstring that the render is *not* bit-exact.
  One page has since been observed three bytes different across two renders of the same inputs.
  ~~The target still reports DIFFERS, for one reason only: `tools/shrink-pngs` runs ImageMagick, which stamps three `date:create` / `date:modify` / `date:timestamp` tEXt chunks with the wall clock.~~
  **Fixed, and it needed no ruling from Aria:** `shrink-pngs` passes `-define png:exclude-chunk=date`, and the shipped pages carry no `tEXt`, `tIME` or `iTXt` chunk at all (checked 2026-09-26 at r268).
  The repo-weight concern this bullet opened has moved rather than gone: the atlas does not churn on a no-op render, and it has been *deliberately* re-cut 34 times, which is item 58.
- `proto` (11%), `app` (16%) and `render` (5%) are the low-coverage packages; `proto` is exercised through the daemon tests, `render` is the GUI.
- `botropolis-notify` is installed but not enabled; check `pacman -Q botropolis-git` against the PKGBUILD before validating.
- **`log.showSignature=true` is set on this repo, so `git log` prints a verification line per commit.** `git log --oneline | wc -l` counts double and `git log --format=%s` returns the signature before the subject.
  Pass `--no-show-signature` in anything scripted; `git rev-list --count` and `git rev-parse` are unaffected, which is why `pkgver()` never broke.
  It has produced two wrong results in one session — a commit count in item 58 and a package subject in item 57's Makefile — so it is a property of the repo rather than a mistake either of us made twice.

### Validation checklist for Aria

Install the latest build, restart the daemon, enable notify, then:

1. `botropolis status` — every live session's state matches what you know it is doing (bug 1 and 2 will show here).
2. `botropolis` — Tab to the first needs-you, Enter: terminator opens with `claude attach` on that session, and Ctrl-Z leaves it running.
3. `c` on a district: a new background session in that folder, and it appears on the map within a few seconds.
4. Let a session hand a turn back while the map is unfocused: the desktop notification fires and the away panel shows it on refocus.
5. `x` on the plant: the breakdown's 24 h cost is within a few dollars of `~$… 24h` on the strip, and no row says `$0.00` for millions of tokens.
6. `b`, `t`, `s`, `?`, `h`, `r`, `n`, `/` — each opens, closes with Escape, and none leaves the map in a wrong state.
7. `d d` on a parked container: it is gone from the map and from `claude agents --json --all`.
8. `botropolis bar` in waybar: the class changes colour when a session needs you.
9. `b` with more projects than fit the window: the sidebar scrolls and the cursor row stays visible.
10. Type `!` in a session to drop into a shell, then look at the map: the session reads as working, not needs-you (bug 12).

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
