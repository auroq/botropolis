const REQUIRED = ['name', 'spacing', 'group'];

export class VarietyError extends Error {}

export function normaliseVariety(raw, index = 0) {
  for (const key of REQUIRED) {
    if (raw[key] === undefined || raw[key] === '') {
      throw new VarietyError(`variety #${index + 1}: missing "${key}"`);
    }
  }
  const spacing = Number(raw.spacing);
  if (!Number.isFinite(spacing) || spacing <= 0) {
    throw new VarietyError(`variety "${raw.name}": spacing must be a positive number of metres`);
  }
  return {
    name: String(raw.name),
    spacing,
    group: String(raw.group).toUpperCase(),
    selfFertile: Boolean(raw.selfFertile),
    count: raw.count === undefined ? null : Math.max(0, Math.trunc(raw.count)),
  };
}

const ORDER = 'ABCDEFG';

export function groupsCompatible(a, b) {
  const i = ORDER.indexOf(a);
  const j = ORDER.indexOf(b);
  if (i === -1 || j === -1) return false;
  return Math.abs(i - j) <= 1;
}

export function canPollinate(donor, recipient) {
  if (donor.name === recipient.name) return recipient.selfFertile;
  return groupsCompatible(donor.group, recipient.group);
}
