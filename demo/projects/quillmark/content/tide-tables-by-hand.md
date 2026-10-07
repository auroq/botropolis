---
title: Reading tide tables by hand
date: 2026-09-02
author: M. Trevose
---

Before the app, there was the almanac.
A tide table lists the times and heights of high and low water for one port.

## The rule of twelfths

Between low and high water the tide does not rise evenly:

1. first hour: one twelfth
2. second hour: two twelfths
3. third hour: three twelfths
4. fourth hour: three twelfths
5. fifth hour: two twelfths
6. sixth hour: one twelfth

In code it is a short loop:

```python
def twelfths(low, high):
    step = (high - low) / 12
    return [low + step * n for n in (0, 1, 3, 6, 9, 11, 12)]
```

Multiply by the *range*, not the height, or the answer drifts.
