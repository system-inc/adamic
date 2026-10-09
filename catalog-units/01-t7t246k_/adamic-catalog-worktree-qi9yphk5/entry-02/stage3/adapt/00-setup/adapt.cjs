// Setup delegates to upstream rather than rewriting compiler source.
const { resolve } = require('node:path');
const { spawnSync } = require('node:child_process');
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = resolve(process.argv[2]);
const result = spawnSync(process.execPath, ['scripts/processDiagnosticMessages.mjs', 'src/compiler/diagnosticMessages.json'], { cwd: tree, stdio: 'inherit' });
if (result.error) throw result.error;
process.exit(result.status ?? 1);
