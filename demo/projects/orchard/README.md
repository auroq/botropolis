# orchard

Plan an orchard before you dig.
Give it a plot and a list of tree varieties, and it lays the trees out in rows and checks that every tree has a pollination partner close enough to matter.

It has no dependencies and needs Node 20 or later.

## Usage

```sh
node bin/orchard.js examples/hillside.json
node bin/orchard.js --json --range 12 examples/lonely-pear.json
```

The CLI exits 1 when any tree is left without a partner, so it can sit in a script.

## A spec

```json
{
  "plot": { "width": 40, "length": 30, "rowGap": 5, "margin": 2 },
  "pollinationRange": 15,
  "varieties": [
    { "name": "Egremont Russet", "spacing": 4, "group": "C" },
    { "name": "Bramley", "spacing": 6, "group": "D", "count": 6 }
  ]
}
```

Rows run across the width and step down the length, one variety per row, cycling through the list.
A variety with a `count` stops being planted once that many trees are placed.

## Pollination groups

Varieties flower in groups `A` to `G`, earliest first.
Two trees can pollinate each other when their groups are the same or adjacent.

## As a library

```js
import { plan, formatPlan } from 'orchard';

console.log(formatPlan(plan(spec)));
```

## Tests

```sh
npm test
```
