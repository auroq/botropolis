import { layoutRows } from './layout.js';
import { checkPollination, DEFAULT_RANGE } from './pollination.js';
import { normaliseVariety } from './varieties.js';

export function plan(spec) {
  const varieties = (spec.varieties ?? []).map(normaliseVariety);
  const rows = layoutRows(spec.plot ?? {}, varieties);
  const range = spec.pollinationRange ?? DEFAULT_RANGE;
  const problems = checkPollination(rows, varieties, range);
  const counts = {};
  for (const row of rows) counts[row.variety] = (counts[row.variety] ?? 0) + row.trees.length;
  return { rows, problems, counts, total: Object.values(counts).reduce((a, b) => a + b, 0) };
}

export function formatPlan(result) {
  const lines = [];
  for (const row of result.rows) {
    lines.push(`row ${String(row.index + 1).padStart(2)}  y=${row.y.toFixed(1).padStart(5)} m  ${row.variety} x${row.trees.length}`);
  }
  lines.push('');
  for (const [name, n] of Object.entries(result.counts)) lines.push(`${name}: ${n}`);
  lines.push(`total: ${result.total} trees`);
  if (result.problems.length === 0) {
    lines.push('pollination: ok');
  } else {
    lines.push(`pollination: ${result.problems.length} tree(s) without a partner`);
    for (const p of result.problems) {
      lines.push(`  ${p.tree.variety} at (${p.tree.x}, ${p.tree.y}): ${p.reason}`);
    }
  }
  return lines.join('\n');
}
