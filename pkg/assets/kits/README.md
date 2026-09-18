# Atlases rendered from Kenney's 3D kits

The `kits-z*.png` pages and their `kits-z*.json` manifests are rendered by `tools/render-sprites/render.py`
(Blender 5.2, headless) from [Kenney](https://kenney.nl)'s CC0 3D kits, fetched by `make kits` into the gitignored `tools/kits/`.
Only the rendered pages, the manifests and each kit's `License.txt` are committed.
`make sprites` regenerates them; the page budget is eight 2048 px pages per zoom level.

| Kit | Page | Used for |
| --- | --- | --- |
| City Kit Commercial | https://kenney.nl/assets/city-kit-commercial | session buildings by context fill; the city hall and library |
| City Kit Roads | https://kenney.nl/assets/city-kit-roads | avenues, bends, crossings, ends; lamp posts |
| City Kit Industrial | https://kenney.nl/assets/city-kit-industrial | the power plant and its stack; the towers (water towers); parked sessions as shipping containers |
| City Kit Suburban | https://kenney.nl/assets/city-kit-suburban | the trees in the parks and the belt |
| Car Kit | https://kenney.nl/assets/car-kit | the cars on avenues with traffic (scaled to a third of a cell) |
| Space Kit | https://kenney.nl/assets/space-kit | the rover at the door of a working session |
| (modelled in the script) | — | the drone that stands for a subagent in flight |

Every piece is cut at four headings (0, 90, 180, 270 degrees) under one sun,
seen by an orthographic camera tilted atan(1/2) above the ground and turned 45 degrees,
so a one-unit tile projects as the map's 2:1 diamond, 132 px wide at zoom 1.
Each sprite's manifest entry gives its page, its rectangle and the pixel where the piece's ground origin lands.

Kenney's terms, from [kenney.nl/support](https://kenney.nl/support): CC0, free for any use, attribution welcome but not required.
Thank you, Kenney; [support the studio](https://kenney.nl/donate) if the work helps you too.
