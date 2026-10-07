import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { layoutRows, PlotError } from '../src/layout.js';
import { normaliseVariety } from '../src/varieties.js';

const variety = (name, spacing, extra = {}) => normaliseVariety({ name, spacing, group: 'C', ...extra });

describe('layoutRows', () => {
  describe('when two varieties share a plot', () => {
    const rows = layoutRows({ width: 10, length: 10, rowGap: 4, margin: 1 }, [variety('a', 2), variety('b', 4)]);

    it('should alternate varieties row by row', () => {
      assert.deepEqual(rows.map((r) => r.variety), ['a', 'b', 'a']);
    });
    it('should place trees at the variety spacing', () => {
      assert.deepEqual(rows[1].trees.map((t) => t.x), [1, 5, 9]);
    });
    it('should space rows by the larger of rowGap and spacing', () => {
      assert.deepEqual(rows.map((r) => r.y), [1, 5, 9]);
    });
  });

  describe('when a variety has a count', () => {
    const rows = layoutRows({ width: 20, length: 20, margin: 0 }, [variety('rare', 2, { count: 3 }), variety('common', 5)]);
    const rare = rows.filter((r) => r.variety === 'rare').flatMap((r) => r.trees);

    it('should place no more than the count', () => assert.equal(rare.length, 3));
    it('should keep filling with the other varieties', () => {
      assert.ok(rows.slice(2).every((r) => r.variety === 'common'));
    });
  });

  describe('when there are no varieties', () => {
    it('should return no rows', () => assert.deepEqual(layoutRows({ width: 5, length: 5 }, []), []));
  });

  describe('when the plot has no area', () => {
    it('should throw a PlotError', () => assert.throws(() => layoutRows({ width: 0, length: 5 }, [variety('a', 1)]), PlotError));
  });
});
