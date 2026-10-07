export class PlotError extends Error {}

export function layoutRows(plot, varieties) {
  const { width, length, rowGap = 4, margin = 1 } = plot;
  if (!(width > 0) || !(length > 0)) {
    throw new PlotError('plot width and length must be positive');
  }
  if (varieties.length === 0) return [];

  const usableWidth = width - margin * 2;
  const usableLength = length - margin * 2;
  const rows = [];
  let y = margin;
  let v = 0;

  while (y <= margin + usableLength) {
    const variety = pickNext(varieties, v);
    if (!variety) break;
    const trees = [];
    for (let x = margin; x <= margin + usableWidth; x += variety.spacing) {
      if (variety.count !== null && variety.placed >= variety.count) break;
      trees.push({ variety: variety.name, x, y });
      variety.placed += 1;
    }
    if (trees.length > 0) rows.push({ index: rows.length, variety: variety.name, y, trees });
    y += Math.max(rowGap, variety.spacing);
    v += 1;
  }
  return rows;
}

function pickNext(varieties, start) {
  for (const v of varieties) v.placed ??= 0;
  for (let k = 0; k < varieties.length; k += 1) {
    const candidate = varieties[(start + k) % varieties.length];
    if (candidate.count === null || candidate.placed < candidate.count) return candidate;
  }
  return null;
}
