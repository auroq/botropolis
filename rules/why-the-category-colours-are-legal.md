## Why the category colours are legal

The rule everywhere else is that one colour means one thing: the state tones are the same in the map, the strip, the cards, the TUI and the waybar class.
The info views break it — a view tints by its own ramp or its own categories — and the exception was granted on four conditions, of which the fourth was that the category colours stay clear of the state tones.

**That condition cannot be met by choosing hues.**
Of the 56 ways to take three of the reference palette's eight dark slots, 15 pass the all-pairs gates a map needs; every one of the 15 comes within ΔE 2.2 of some state tone under one of the deficiencies, and the best of them is the set in use (`#3987e5` blue, `#d95926` orange, `#199e70` aqua, colliding with unattended violet under deuteranopia at 2.2).
The sequential ramp is no better placed: walked in 41 steps it passes within 0.3 of the waiting teal under deuteranopia.
Seven state tones and three categories do not fit in the space the dark surface leaves once the lightness band and the chroma floor are applied.
Checked independently under normal vision: blue sits 6.2 from working, orange 5.0 from error, aqua 10.5 from waiting — all under the adjacency floor of 15, which is only survivable because they are never adjacent.

**What keeps the condition is the mode, not the palette.**
A view is subtractive: while one is up the map is receded and tinted by the view alone, and the strip gives its tones back.
A category colour and a state tone are never on screen together, so the ambiguity the condition guards against cannot arise.

The consequence is worth stating plainly, because it is a coupling and not a preference: **"the strip recedes" is load-bearing.**
If the state tones ever come back while a view is up, the palette is illegal again and no choice of hue repairs it.
`TestDrained` in `pkg/ui/strip_test.go` is what holds it — "it should leave no state tone anywhere on the strip" — and it must not be relaxed without re-opening this section.
