#!/usr/bin/env node
import { readFile } from 'node:fs/promises';
import { parseArgs } from 'node:util';
import { formatPlan, plan } from '../src/index.js';

const USAGE = `usage: orchard [--json] [--range METRES] <spec.json>

Lay out an orchard from a JSON spec and check pollination.

  --json          print the plan as JSON instead of text
  --range METRES  override the pollination range in the spec
  -h, --help      show this help`;

async function main(argv) {
  const { values, positionals } = parseArgs({
    args: argv,
    allowPositionals: true,
    options: {
      json: { type: 'boolean', default: false },
      range: { type: 'string' },
      help: { type: 'boolean', short: 'h', default: false },
    },
  });
  if (values.help) {
    console.log(USAGE);
    return 0;
  }
  if (positionals.length !== 1) {
    console.error(USAGE);
    return 2;
  }
  const spec = JSON.parse(await readFile(positionals[0], 'utf8'));
  if (values.range !== undefined) spec.pollinationRange = Number(values.range);
  const result = plan(spec);
  console.log(values.json ? JSON.stringify(result, null, 2) : formatPlan(result));
  return result.problems.length === 0 ? 0 : 1;
}

main(process.argv.slice(2)).then(
  (code) => process.exit(code),
  (err) => {
    console.error(`orchard: ${err.message}`);
    process.exit(2);
  },
);
