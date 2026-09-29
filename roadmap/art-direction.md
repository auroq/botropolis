## 3. Art direction: decided 2026-09-18

The current look reads as Age of Empires or early SimCity rather than Factorio, and the reason is not 2D versus 3D.
Factorio reads as deliberate because every sprite shares one light direction with a cast shadow, one palette grade, one tile scale,
and detail that is dense but ordered.
The current map mixes three pack families at two scales, has no shadows, sits on saturated noise, and spaces its plots out.
The fix is a committed style, and the cheapest way to get one is Factorio's own method: model in 3D, render once, ship 2D.

**Runtime stays 2D and stays Go.**
The 3D happens offline in Blender at build time; Ebitengine keeps drawing y-sorted sprites from an atlas exactly as it does now.
Blender (5.2, in `extra`) is a developer dependency for regenerating atlases; the committed PNGs mean the package and the runtime never touch it.
Free tilt is the one thing this path cannot give — only the four fixed headings — and it is the only thing that would reopen the engine question.

**Packs, all CC0, all Kenney unless noted, chosen for one shared palette (slate, off-white, one amber accent — which is also the needs-you colour):**

| Role | Pack | Notes |
| --- | --- | --- |
| Buildings, landmarks | [City Kit Commercial](https://kenney.nl/assets/city-kit-commercial), [Industrial](https://kenney.nl/assets/city-kit-industrial), [Suburban](https://kenney.nl/assets/city-kit-suburban) | Industrial has smokestacks, a water tower, silos and a cooling tower: the plant, the towers and the hall have native pieces |
| Streets, avenues, plaza | [City Kit Roads](https://kenney.nl/assets/city-kit-roads) | |
| Parks, greenbelt | [City Kit Suburban](https://kenney.nl/assets/city-kit-suburban) trees | placed by the plan with seeded in-cell offsets; the Nature Kit was tried and dropped 2026-09-18 — its teal trees are taller than the buildings and off the palette |
| Workers (main thread) | [Space Kit](https://kenney.nl/assets/space-kit) rovers | same author and palette; motion is a bob and a wheel spin from the pipeline |
| Subagents in flight | our own drone, modelled in the pipeline (a body, two rotors, one accent light) | Factorio's logistic bot is exactly this; guarantees the palette and the state light with no third-party asset |
| Traffic | [Car Kit](https://kenney.nl/assets/car-kit) | cars per road in proportion to traffic |
| Spend (the ledger) | [Train Kit](https://kenney.nl/assets/train-kit) | a rail line from the plant to each district: a train per model, wagons per thousand tokens over the breakdown window; decided 2026-09-18 |
| Live rate and telemetry (the current) | power poles and wires, as built | sparks at tokens/min as today; **no wire means the daemon has seen no hook events for that session and is reading files** — the one datum nothing showed |
| Tower names | signage on the building | short names horizontal on the face, long names vertical up the side, or a billboard on the roof, like a company name on an office block; never a floating plate |
| River: arrivals and departures | [Watercraft Kit](https://kenney.nl/assets/watercraft-kit) | decided 2026-09-18: the river carries the session lifecycle — a new session arrives on a barge and docks at its district before its building rises, a demolished session's container leaves downriver. Fallback if that is too much motion: one or two slow boats as scenery, the card saying so |
| Held in reserve | [Quaternius Animated Robot Pack](https://quaternius.com/packs/animatedrobot.html) (CC0) | a rigged robot with walk and idle clips, if the workers ever want personality rather than machinery |

Workers are bots, not characters.
That is the line between Factorio and Animal Crossing, and the whole point of the redesign is to be on the Factorio side of it.

Kenney has no robot or drone character pack; the catalogue was checked, not remembered.
