# 48. Cars belong on the main view

**Done 2026-09-26.**

One line: `shown` in `pkg/city/network.go` had no entry for `ViewAttention`, so the base view drew no networks. `ViewAttention: {NetworkTraffic}` puts the cars back.

Recorded against it, because it is a real tension rather than a caveat: the info-view model is subtractive on purpose, so that a view can say something by taking things away. On the traffic view a car is explained by the view it is in; on the base view it is unexplained motion, which is the exact complaint that started this project.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
