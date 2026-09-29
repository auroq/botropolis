# 64. The tick gate does not know about the courier

**Done 2026-09-26 (`546854e`, r283).**

Found 2026-09-26 reading `Animating()` while measuring item 60, and the build session independently asked for it to be filed.

`pkg/render/watching.go:83` sets the tick to `liveTPS` when the scene is animating and `stillTPS` — 12 — when it is not.
`city.Scene.Animating()` lists voyages, trains, streets with traffic, and buildings that pulse, smoke, build or work.
**The refresh courier is none of those.** It lives in its own slice (`Scene.couriers`, `courier.go:73`), and nothing in the gate reads it.

So on a quiet city — nothing working, no traffic, no train — pressing `u` launches a boat that crosses the entire river in five seconds while the loop runs at 12 fps: about 60 frames for the traverse instead of 150.
That is the fastest-moving object on the map running at the slowest tick, and item 56's whole case for the courier is that the motion is the explanation.

Two things, and the second is bookkeeping that keeps the first from coming back:
1. Add the couriers to the gate. `Couriers()` already resolves them per frame, so it is `len(s.couriers) > 0` with the same live-resolution rule. The test is a scene with nothing else moving: `SendCourier`, then `Animating()` must be true.
2. **Record the gauge boats as a deliberate exclusion**, in the paragraph that already names the fountain and the lamps. They belong there for a good reason — `gaugeAt` takes no clock, so a gauge's position *is* its reading and it is repainted in place — and right now their absence looks exactly like the courier's, which is an oversight. A list whose comment says it names "every object that moves between frames and nothing else" needs the exceptions written down, or the next reader cannot tell a decision from a gap.

**Fixed, both halves.**

The gate asks the clock rather than the slice length: `for _, courier := range s.couriers { if courier.Progress(s.now()) < 1 { return true } }`.
Not `len(s.couriers) > 0` as filed, and not `len(s.Couriers()) > 0` either.
`Couriers()` prunes *and* allocates a copy for the renderer, which is more than a gate running every `Update` should do, and the bare length would hold the tick open after the boat had landed, because pruning only happens when the renderer asks.
Asking each courier's own progress is the same live-resolution rule with neither cost.

**Two tests, and the second one is the one that needed proving.**
The first is the filed case: a city of parked sessions, `require.False(Animating())` first so the scene is known still, then `SendCourier`, then it must animate.
The second winds the clock past `CourierFor` and asserts the city goes still again — and *that* test passed before the fix, because `Animating()` never returned true for a courier at all.
So I mutated the fix to the filed `len(s.couriers) > 0` and confirmed the expiry test fails against it, which is the only thing that distinguishes a test that holds the behaviour from one that is merely green.

The exclusions paragraph now names three things and says why the list needs them written down at all: the fountain's three spray frames, the lamps' glow fixed by zoom and hour, and the gauges, which hold still because `gaugeAt` takes no clock and a boat that drifted between frames would be lying about its reading.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
