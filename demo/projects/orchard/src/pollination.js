import { canPollinate } from './varieties.js';

export const DEFAULT_RANGE = 15;

export function distance(a, b) {
  return Math.hypot(a.x - b.x, a.y - b.y);
}

export function checkPollination(rows, varieties, range = DEFAULT_RANGE) {
  const byName = new Map(varieties.map((v) => [v.name, v]));
  const trees = rows.flatMap((row) => row.trees);
  const problems = [];

  for (const tree of trees) {
    const variety = byName.get(tree.variety);
    if (variety.selfFertile) continue;
    const partner = trees.find(
      (other) =>
        other !== tree &&
        distance(tree, other) <= range &&
        canPollinate(byName.get(other.variety), variety),
    );
    if (!partner) problems.push({ tree, reason: `no group ${variety.group}-compatible pollinator within ${range} m` });
  }
  return problems;
}
