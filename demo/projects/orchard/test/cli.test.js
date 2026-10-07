import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { describe, it } from 'node:test';

const root = fileURLToPath(new URL('..', import.meta.url));
const run = (...args) => spawnSync(process.execPath, ['bin/orchard.js', ...args], { cwd: root, encoding: 'utf8' });

describe('orchard CLI', () => {
  describe('when every tree has a partner', () => {
    const result = run('examples/hillside.json');

    it('should exit 0', () => assert.equal(result.status, 0));
    it('should print the total', () => assert.match(result.stdout, /total: 56 trees/));
  });

  describe('when a tree has no partner', () => {
    const result = run('examples/lonely-pear.json');

    it('should exit 1', () => assert.equal(result.status, 1));
    it('should name the stranded variety', () => assert.match(result.stdout, /Doyenné du Comice at/));
  });

  describe('when asked for JSON', () => {
    it('should print parseable JSON', () => {
      assert.equal(JSON.parse(run('--json', 'examples/hillside.json').stdout).total, 56);
    });
  });

  describe('when the range is overridden', () => {
    it('should use it', () => assert.match(run('--range', '3', 'examples/hillside.json').stdout, /within 3 m/));
  });

  describe('when given no spec', () => {
    it('should exit 2', () => assert.equal(run().status, 2));
  });

  describe('when asked for help', () => {
    it('should print usage', () => assert.match(run('--help').stdout, /^usage: orchard/));
  });
});
