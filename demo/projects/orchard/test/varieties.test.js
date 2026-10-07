import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { canPollinate, groupsCompatible, normaliseVariety, VarietyError } from '../src/varieties.js';

describe('normaliseVariety', () => {
  describe('when given a complete variety', () => {
    const v = normaliseVariety({ name: 'Discovery', spacing: '4', group: 'c', count: 3.7 });

    it('should coerce spacing to a number', () => assert.equal(v.spacing, 4));
    it('should upper-case the group', () => assert.equal(v.group, 'C'));
    it('should truncate the count', () => assert.equal(v.count, 3));
    it('should default selfFertile to false', () => assert.equal(v.selfFertile, false));
  });

  describe('when the count is omitted', () => {
    it('should leave it unlimited', () => {
      assert.equal(normaliseVariety({ name: 'a', spacing: 1, group: 'A' }).count, null);
    });
  });

  const invalid = [
    ['the name is missing', { spacing: 4, group: 'C' }],
    ['the group is empty', { name: 'x', spacing: 4, group: '' }],
    ['the spacing is zero', { name: 'x', spacing: 0, group: 'C' }],
    ['the spacing is not a number', { name: 'x', spacing: 'wide', group: 'C' }],
  ];
  for (const [when, raw] of invalid) {
    describe(`when ${when}`, () => {
      it('should throw a VarietyError', () => assert.throws(() => normaliseVariety(raw), VarietyError));
    });
  }
});

describe('groupsCompatible', () => {
  const cases = [
    ['C', 'C', true],
    ['C', 'D', true],
    ['D', 'C', true],
    ['B', 'D', false],
    ['C', 'Z', false],
  ];
  for (const [a, b, expected] of cases) {
    describe(`when comparing ${a} and ${b}`, () => {
      it(`should return ${expected}`, () => assert.equal(groupsCompatible(a, b), expected));
    });
  }
});

describe('canPollinate', () => {
  const selfish = normaliseVariety({ name: 'Conference', spacing: 5, group: 'C', selfFertile: true });
  const needy = normaliseVariety({ name: 'Comice', spacing: 5, group: 'E' });

  describe('when the donor is the same self-fertile variety', () => {
    it('should return true', () => assert.equal(canPollinate(selfish, selfish), true));
  });
  describe('when the donor is the same self-sterile variety', () => {
    it('should return false', () => assert.equal(canPollinate(needy, needy), false));
  });
  describe('when the groups are two apart', () => {
    it('should return false', () => assert.equal(canPollinate(selfish, needy), false));
  });
});
