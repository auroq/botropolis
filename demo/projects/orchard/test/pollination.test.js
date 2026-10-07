import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { checkPollination, distance } from '../src/pollination.js';
import { normaliseVariety } from '../src/varieties.js';

const tree = (variety, x, y) => ({ variety, x, y });
const rowsOf = (...trees) => [{ index: 0, variety: trees[0].variety, y: 0, trees }];

describe('distance', () => {
  it('should be euclidean', () => assert.equal(distance({ x: 0, y: 0 }, { x: 3, y: 4 }), 5));
});

describe('checkPollination', () => {
  const varieties = [
    normaliseVariety({ name: 'Russet', spacing: 4, group: 'C' }),
    normaliseVariety({ name: 'Kernel', spacing: 4, group: 'D' }),
    normaliseVariety({ name: 'Comice', spacing: 4, group: 'E' }),
    normaliseVariety({ name: 'Conference', spacing: 4, group: 'C', selfFertile: true }),
  ];

  describe('when compatible partners are within range', () => {
    it('should report no problems', () => {
      const rows = rowsOf(tree('Russet', 0, 0), tree('Kernel', 10, 0));
      assert.deepEqual(checkPollination(rows, varieties, 15), []);
    });
  });

  describe('when the only partner is out of range', () => {
    const rows = rowsOf(tree('Russet', 0, 0), tree('Kernel', 20, 0));
    const problems = checkPollination(rows, varieties, 15);

    it('should report both trees', () => assert.equal(problems.length, 2));
    it('should name the range in the reason', () => assert.match(problems[0].reason, /within 15 m/));
  });

  describe('when the neighbour is the same self-sterile variety', () => {
    it('should report both trees', () => {
      const rows = rowsOf(tree('Russet', 0, 0), tree('Russet', 4, 0));
      assert.equal(checkPollination(rows, varieties, 15).length, 2);
    });
  });

  describe('when a tree is self-fertile and alone', () => {
    it('should report no problems', () => {
      assert.deepEqual(checkPollination(rowsOf(tree('Conference', 0, 0)), varieties, 15), []);
    });
  });

  describe('when the groups are two apart', () => {
    it('should report both trees', () => {
      const rows = rowsOf(tree('Russet', 0, 0), tree('Comice', 4, 0));
      assert.equal(checkPollination(rows, varieties, 15).length, 2);
    });
  });
});
