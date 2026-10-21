#!/usr/bin/env node
// Isolated SDK wrapper: release flags apply to runtime archives as well as entry C.
import { readFileSync, appendFileSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
const config = JSON.parse(readFileSync(new URL('./configuration.json', import.meta.url)));
const args = process.argv.slice(2).map(value => value === '-O2' ? '-Oz' : value);
function run(argv) {
  appendFileSync(config.commands, JSON.stringify(argv) + '\n');
  const result = spawnSync(argv[0], argv.slice(1), { stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}
if (args.length === 1 && args[0] === '--version') {
  run([config.clang, '--version']);
  console.log('handler release configuration: ' + JSON.stringify(config));
} else {
  run([config.clang, ...args]);
  if (!args.includes('-c') && args.includes('-o')) {
    const output = args[args.indexOf('-o') + 1];
    run([config.strip, output]);
    run([config.wasmOpt, '-Oz', output, '-o', output]);
  }
}
